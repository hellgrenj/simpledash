import { connectAndConsume } from "./ws.js";
const App = {
    element: document.getElementById('app'),
    simpledashContext: {},
    $: {
        getContext: async () => {
            const response = await fetch('/context');
            App.simpledashContext = await response.json();
        },
        renderHeader: () => {
            const header = document.createElement('div');
            header.className = 'header';
            header.innerHTML = `<h2>simpledash v.0.1 - Cluster: ${App.simpledashContext.ClusterName}</h2>`;
            App.element.appendChild(header);
        },
        addIngressSection: () => {
            const ingress = document.createElement('div');
            ingress.className = 'ingress';
            App.element.appendChild(ingress);
        },
        namespaceColors: {},
        getRandomColor: () => {
            var letters = '0123456789ABCDEF';
            var color = '#';
            for (var i = 0; i < 6; i++) {
                color += letters[Math.floor(Math.random() * 16)];
            }
            return color;
        },
        createNode: (nodeKey, nodeIndex) => {
            const node = document.createElement('div');
            node.className = 'node';
            node.style = `grid-column: ${nodeIndex};`;
            node.innerHTML = `<div><span>${new Date().toLocaleTimeString()}</span><h2>${nodeKey}</h2>(hover over pod to se image)<br/></div>`;
            return node;
        },
        clearNodes: () => {
            const elements = document.getElementsByClassName('node');
            while (elements.length > 0) {
                elements[0].parentNode.removeChild(elements[0]);
            }
        },
        createPod: (pod) => {
            let podBgColor = "#FFFFFF";
            if (pod.Status == "Running") {
                podBgColor = "rgb(182, 218, 129)";
            } else if (pod.Status == "Failed") {
                podBgColor = "rgb(148, 61, 61)";
            } else {
                podBgColor = "rgb(202, 202, 110)"
            }
            const podElement = document.createElement('div');
            if (App.$.namespaceColors[pod.Namespace] === undefined) {
                App.$.namespaceColors[pod.Namespace] = App.$.getRandomColor();
            }
            podElement.className = 'pod';
            podElement.style = `background-color: ${podBgColor}; border: 4px solid ${App.$.namespaceColors[pod.Namespace]};`;
            podElement.innerHTML = `
                   ${pod.Name}<br/>
                   ${pod.Namespace}
                `;
            podElement.title = pod.Image;
            return podElement;
        },
        renderNode: (clusterInfo, nodeElement, key) => {
            clusterInfo.Nodes[key].forEach((pod) => {
                const podElement = App.$.createPod(pod)
                nodeElement.appendChild(podElement);
            });
        },
        renderIngress: (clusterInfo) => {
            const ingressElement = document.getElementsByClassName('ingress')[0];
            let ingressHtml = 'Endpoints';
            if (clusterInfo.Ingresses) {
                clusterInfo.Ingresses.forEach(ingress => {
                    ingressHtml = `${ingressHtml} <br/> ${ingress.Endpoint}`;
                });
            }
            ingressElement.innerHTML = ingressHtml;
        }
    },
    init: async () => {
        await App.$.getContext();
        App.$.renderHeader();
        App.$.addIngressSection();
        connectAndConsume((e) => {
            const clusterInfo = JSON.parse(e.data);
            App.$.clearNodes();
            Object.keys(clusterInfo.Nodes).forEach(async (key, nodeIndex) => {
                const nodeElement = App.$.createNode(key, nodeIndex);
                App.$.renderNode(clusterInfo, nodeElement, key);
                App.$.renderIngress(clusterInfo);
                App.element.appendChild(nodeElement);
            })
        });
    }
}
App.init();

