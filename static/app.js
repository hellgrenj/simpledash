import { connectAndConsume } from "./ws.js";
const App = {
    element: document.getElementById('app'),
    simpledashContext: {},
    state: {
        nsfilter: '',
        clusterInfo: {},
        namespaceColors: {},
        latestTimeStamp: '',
    },
    $: {
        getContext: async () => {
            const response = await fetch('/context');
            App.simpledashContext = await response.json();
        },
        renderHeader: () => {
            const header = document.createElement('div');
            header.className = 'header';
            const html = `
            <h2>simpledash v.0.1 - Cluster: ${App.simpledashContext.ClusterName}</h2>
            <input id="nsfilter" placeholder="filter on namespace">
            `
            header.innerHTML = html;
            App.element.appendChild(header);
            const nsfilterInput = document.getElementById("nsfilter");
            nsfilterInput.addEventListener('keyup', () => {
                App.state.nsfilter = nsfilterInput.value;
                App.$.renderClusterInfo(App.state.latestTimeStamp);
            });
        },
        addIngressSection: () => {
            const ingress = document.createElement('div');
            ingress.className = 'ingress';
            App.element.appendChild(ingress);
        },
        getRandomColor: () => {
            var letters = '0123456789ABCDEF';
            var color = '#';
            for (var i = 0; i < 6; i++) {
                color += letters[Math.floor(Math.random() * 16)];
            }
            return color;
        },
        createNode: (nodeKey, nodeIndex, timeString) => {
            const node = document.createElement('div');
            node.className = 'node';
            node.style = `grid-column: ${nodeIndex};`;
            node.innerHTML = `<div><span>${timeString}</span><h2>${nodeKey}</h2>(hover over pod to see image)<br/></div>`;
            return node;
        },
        clearNodes: () => {
            const elements = document.getElementsByClassName('node');
            while (elements.length > 0) {
                elements[0].parentNode.removeChild(elements[0]);
            }
        },
        createPod: (pod) => {
            if (App.state.nsfilter !== '' && !pod.Namespace.startsWith(App.state.nsfilter)) {
                return null;
            }
            let podBgColor = "#FFFFFF";
            if (pod.Status == "Running") {
                podBgColor = "rgb(182, 218, 129)";
            } else if (pod.Status == "Failed") {
                podBgColor = "rgb(148, 61, 61)";
            } else {
                podBgColor = "rgb(202, 202, 110)"
            }
            const podElement = document.createElement('div');
            if (App.state.namespaceColors[pod.Namespace] === undefined) {
                App.state.namespaceColors[pod.Namespace] = App.$.getRandomColor();
            }
            podElement.className = 'pod';
            podElement.style = `background-color: ${podBgColor}; border: 4px solid ${App.state.namespaceColors[pod.Namespace]};`;
            const imageParts = pod.Image.split(':');
            const html = `
            ${pod.Name}<br/>
            <i>${pod.Namespace}</i><br/>
            <b>tag: ${imageParts[imageParts.length - 1]}</b>
         `
            podElement.innerHTML = html;
            podElement.title = pod.Image;
            return podElement;
        },
        renderNode: (nodeElement, key) => {
            App.state.clusterInfo.Nodes[key].forEach((pod) => {
                const podElement = App.$.createPod(pod)
                if (podElement) {
                    nodeElement.appendChild(podElement);
                }
            });
        },
        renderIngress: () => {
            const ingressElement = document.getElementsByClassName('ingress')[0];
            let ingressHtml = 'Endpoints: <br/>';
            if (App.state.clusterInfo.Ingresses) {
                App.state.clusterInfo.Ingresses.forEach(ingress => {
                    ingressHtml = `${ingressHtml} <br/> ${ingress.Endpoint} (${ingress.Ip})`;
                });
            }
            ingressElement.innerHTML = ingressHtml;
        },
        renderClusterInfo: (timeString) => {
            App.$.clearNodes();
            Object.keys(App.state.clusterInfo.Nodes).forEach(async (key, nodeIndex) => {
                const nodeElement = App.$.createNode(key, nodeIndex, timeString);
                App.$.renderNode(nodeElement, key);
                App.$.renderIngress();
                App.element.appendChild(nodeElement);
            })
        }
    },
    init: async () => {
        await App.$.getContext();
        App.$.renderHeader();
        App.$.addIngressSection();
        connectAndConsume((e) => {
            const clusterInfo = JSON.parse(e.data);
            App.state.clusterInfo = clusterInfo;
            App.state.latestTimeStamp = new Date().toLocaleTimeString();
            App.$.renderClusterInfo(App.state.latestTimeStamp);
        });
    }
}
App.init();

