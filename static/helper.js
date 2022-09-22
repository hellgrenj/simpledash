export const saveToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(function () {
        console.log(`successfully saved ${text} to the clipboard`);
    }, function () {
        console.error(`failed to save ${text} to the clipboard`);
    });
};
export const getRandomColor = () => {
    var letters = '0123456789ABCDEF';
    var color = '#';
    for (var i = 0; i < 6; i++) {
        color += letters[Math.floor(Math.random() * 16)];
    }
    return color;
};