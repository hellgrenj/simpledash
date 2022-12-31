import { connectAndConsume } from "./ws.js";
import { saveToClipboard, stringToColour } from "./helper.js";

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
        renderHeader: () => {
            const header = document.createElement('div');
            header.className = 'header';
            const html = `
            <h2>simpledash - ${App.simpledashContext.ClusterName}</h2>
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
        addDeploymentsSection: () => {
            const deployments = document.createElement('div');
            deployments.className = 'deployments';
            App.element.appendChild(deployments);
        },
        addNamespaceSection: () => {
            const namespaces = document.createElement('div');
            namespaces.className = 'namespaces';
            App.element.appendChild(namespaces);
        },
        renderNamespaces: () => {
            const namespaces = document.getElementsByClassName('namespaces')[0];
            let html = 'Namespaces: <br/>'  
            App.simpledashContext.Namespaces.forEach(ns => {
                if (App.state.nsfilter !== '' && !ns.startsWith(App.state.nsfilter)) {
                    return;
                }
                html = `${html}<br/><span style="background-color:#000; padding:2px; color: ${App.getColorByNamespace(ns)}">${ns}</span>`
            });
            namespaces.innerHTML = html;
            App.element.appendChild(namespaces);
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
            podElement.className = 'pod';
            podElement.style = `background-color: ${podBgColor}; border: 4px solid ${App.getColorByNamespace(pod.Namespace)};`;
            const imageParts = pod.Image.split(':');
            let html = `
            ${pod.Name}<br/>
            <i>namespace: ${pod.Namespace}</i><br/>
            <span class='tag'>tag: ${imageParts[imageParts.length - 1]}</span><br/>
            status: ${pod.Status}<br/>`
            if (App.simpledashContext.PodLogsLinkEnabled) {
                const podLogsLink = App.simpledashContext.PodLogsLink.replace("PODNAME_PLACEHOLDER", pod.Name);
                html = `${html}<a class="podLogsLink" href="${podLogsLink}" target="_blank">view logs</a>`
            }
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
                    ingressHtml = `${ingressHtml} <br/><a class="endpointLink" href="https://${ingress.Endpoint}" target="_blank" style="color: ${App.getColorByNamespace(ingress.Namespace)};"> ${ingress.Endpoint} (${ingress.Ip})</a>`;
                });
            }
            ingressElement.innerHTML = ingressHtml;
        },
        renderDeployments: () => {
            const deploymentsElement = document.getElementsByClassName('deployments')[0];
            let deploymentsHtml = 'Deployments: <br/>';
            if (App.state.clusterInfo.Deployments) {
                App.state.clusterInfo.Deployments.forEach(deployment => {
                    if (App.state.nsfilter !== '' && !deployment.Namespace.startsWith(App.state.nsfilter)) {
                        return;
                    }
                    deploymentsHtml = `${deploymentsHtml} <br/> <span style="background-color:#000; padding:2px; color: ${App.getColorByNamespace(deployment.Namespace)}; padding: 2px;">${deployment.Name} (ready: ${deployment.ReadyReplicas}/${deployment.Replicas})`;
                    if (App.simpledashContext.DeploymentLogsLinkEnabled) {
                        const deploymentLogsLink = App.simpledashContext.DeploymentLogsLink.replace("DEPLOYMENT_NAME_PLACEHOLDER", deployment.Name).replace("DEPLOYMENT_NAMESPACE_PLACEHOLDER", deployment.Namespace);
                        deploymentsHtml = `${deploymentsHtml} <a class="deploymentLogsLink" href="${deploymentLogsLink}" target="_blank">view logs</a>`;
                    }
                });
            }
            deploymentsElement.innerHTML = deploymentsHtml;
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
            App.$.renderDeployments();
            App.$.renderNamespaces();
            Object.keys(App.state.clusterInfo.Nodes).sort().forEach(async (key, nodeIndex) => {
                const nodeElement = App.$.createNode(key, nodeIndex, timeString);
                App.$.renderNode(nodeElement, key);
                App.$.renderIngress();
                App.element.appendChild(nodeElement);
                App.$.wireClusterInfoEvents();
            })
        }
    },
    getContext: async () => {
        const response = await fetch('/context');
        App.simpledashContext = await response.json();
    },
    getColorByNamespace: (ns) => {
        if (App.state.namespaceColors[ns] === undefined) {
            App.state.namespaceColors[ns] = stringToColour(ns);
        }
        return App.state.namespaceColors[ns];
    },
    init: async () => {
        await App.getContext();
        App.$.renderHeader();
        App.$.addIngressSection();
        App.$.addDeploymentsSection();
        App.$.addNamespaceSection();
        connectAndConsume((e) => {
            const clusterInfo = JSON.parse(e.data);
            App.state.clusterInfo = clusterInfo;
            App.state.latestTimeStamp = clusterInfo.Timestamp;
            App.$.renderClusterInfo(App.state.latestTimeStamp);
        });
    }
}
App.init();

