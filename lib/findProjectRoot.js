"use strict";
// Simple version of `find-project-root`
// https://github.com/kirstein/find-project-root/blob/master/index.js
Object.defineProperty(exports, "__esModule", { value: true });
exports.findProjectRoot = void 0;
const fs = require('fs');
const path = require('path');
const MARKERS = ['.git', '.hg'];
const markerExists = (directory) => {
    return MARKERS.some((mark) => {
        const fileMark = path.join(directory, mark);
        return fs.existsSync(fileMark);
    });
};
function findProjectRoot(directory) {
    while (!markerExists(directory)) {
        const parentDirectory = path.resolve(directory, '..');
        if (parentDirectory === directory) {
            break;
        }
        directory = parentDirectory;
    }
    return directory;
}
exports.findProjectRoot = findProjectRoot;
