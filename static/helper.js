export const saveToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(function () {
        console.log(`successfully saved ${text} to the clipboard`);
    }, function () {
        console.error(`failed to save ${text} to the clipboard`);
    });
};
export const stringToColour = (str) => {
    let stringUniqueHash = [...str].reduce((acc, char) => {
        return char.charCodeAt(0) + ((acc << 5) - acc);
    }, 0);
    return `hsl(${stringUniqueHash % 360}, 95%, 55%)`;
};