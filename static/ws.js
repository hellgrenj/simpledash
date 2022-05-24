export const connectAndConsume = (handleMessage) => {
    const ws = new WebSocket("ws://" + window.location.host + "/ws");
    ws.onopen = () => {
        console.log('successfully connected...');
    }
    ws.onclose = () => {
        console.log('connection closed, trying to reconnect in 3 seconds...');
        setTimeout(() => { connectAndConsume(handleMessage) }, 3000);
    };
    ws.onmessage = (e) => handleMessage(e);
};
