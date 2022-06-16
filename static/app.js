import { connectAndConsume } from "./ws.js";
import { saveToClipboard } from "./helper.js";
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
            <h2>simpledash - Cluster: ${App.simpledashContext.ClusterName}</h2>
            <input id="nsfilter" placeholder="filter on namespace">
            `
            header.innerHTML = html;
            App.element.appendChild(header);
            const nsfilterInput = document.getElementById("nsfilter");
            const params = new URLSearchParams(window.location.search)
            if (params.has('ns')) {
                const initNamespace = params.get('ns');
                App.state.nsfilter = initNamespace;
                nsfilterInput.value = initNamespace;
            }
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
        renderNamespaces: () => {
            const namespaces = document.createElement('div');
            namespaces.className = 'namespaces';
            let html = 'Namespaces: <br/>'
            App.simpledashContext.Namespaces.forEach(ns => html = `${html}<br/>${ns}`);
            namespaces.innerHTML = html;
            App.element.appendChild(namespaces);
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
            node.style = `grid-column: ${nodeIndex + 1};`;
            node.innerHTML = `<div><span>${timeString}</span><h2>${nodeKey}</h2>(click on tag to copy to clipboard)<br/></div>`;
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
            <i>namespace: ${pod.Namespace}</i><br/>
            <span class='tag'>tag: ${imageParts[imageParts.length - 1]}</span><br/>
            status: ${pod.Status}<br/>
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
                    if (App.state.nsfilter !== '' && !ingress.Namespace.startsWith(App.state.nsfilter)) {
                        return;
                    }
                    ingressHtml = `${ingressHtml} <br/> ${ingress.Endpoint} (${ingress.Ip})`;
                });
            }
            ingressElement.innerHTML = ingressHtml;
        },
        wireClusterInfoEvents: () => {
            document.querySelectorAll('.tag').forEach(tag => {
                tag.addEventListener('click', event => {
                    saveToClipboard(event.target.textContent.split('tag: ')[1]);
                });
            });
        },
        renderClusterInfo: (timeString) => {
            App.$.clearNodes();
            Object.keys(App.state.clusterInfo.Nodes).sort().forEach(async (key, nodeIndex) => {
                const nodeElement = App.$.createNode(key, nodeIndex, timeString);
                App.$.renderNode(nodeElement, key);
                App.$.renderIngress();
                App.element.appendChild(nodeElement);
                App.$.wireClusterInfoEvents();
            })
        }
    },
    init: async () => {
        await App.$.getContext();
        App.$.renderHeader();
        App.$.addIngressSection();
        App.$.renderNamespaces();
        connectAndConsume((e) => {
            const clusterInfo = JSON.parse(e.data);
            App.state.clusterInfo = clusterInfo;
            App.state.latestTimeStamp = new Date().toLocaleTimeString();
            App.$.renderClusterInfo(App.state.latestTimeStamp);
        });
    }
}
App.init();

