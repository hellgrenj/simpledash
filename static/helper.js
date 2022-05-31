export const saveToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(function () {
        console.log(`successfully saved ${text} to the clipboard`);
    }, function () {
        console.error(`failed to save ${text} to the clipboard`);
    });
};