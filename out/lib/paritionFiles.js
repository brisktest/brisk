"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.splitFiles = void 0;
const child_process_1 = require("child_process");
const glob_1 = __importDefault(require("glob"));
function lineCount(filename) {
    const results = child_process_1.execSync(`wc -c <  ${filename}`);
    return parseInt(results.toString());
}
function splitFilesByCount(fileArray, lineArray, startLineCount, endLineCount) {
    let startIndex = 0;
    while (startLineCount > lineArray[startIndex]) {
        startIndex++;
    }
    let endIndex = startIndex;
    while (lineArray[endIndex] <= endLineCount) {
        endIndex++;
    }
    return fileArray.slice(startIndex, endIndex);
}
/**
 * writes a list of files matching the glob pattern to stdout
 * runs only the subset of files which fall within the job, set
 * in the environment variables.
 *
 * CI_NODE_TOTAL is the number of jobs we will be splitting across, 1-indexed
 * CI_NODE_INDEX is the index of the job (subset of files) we should be running, 0-indexed
 */
/**
 * gets a list of files matching the given glob
 * @returns {string[]}
 */
function getFiles(numHosts, thisHost, globPattern, cwd) {
    console.debug('the pattern is ', globPattern);
    console.debug('the CWD is ', cwd);
    const allFiles = glob_1.default.sync(globPattern, { cwd });
    const allLines = [...allFiles].map(function (filename) {
        return lineCount(filename);
    });
    const add = (a, b) => a + b;
    const cumulativeLC = allLines.map(function (lc, index) {
        if (index > 0) {
            return allLines.slice(0, index).reduce(add);
        }
        else {
            return allLines[0];
        }
    });
    // now we split the lines into the nodes
    const nodeLines = cumulativeLC[cumulativeLC.length - 1];
    const avLines = Math.ceil(nodeLines / numHosts);
    const startLineCount = avLines * thisHost;
    const endLineCount = startLineCount + avLines;
    return splitFilesByCount(allFiles, cumulativeLC, startLineCount, endLineCount);
}
function splitFiles(numHosts, thisHost, cwd, globPattern = 'tests/**/*.test.ts', splitMethod = 'LINE_COUNT') {
    let files = [];
    switch (splitMethod) {
        case 'LINE_COUNT':
            files = getFiles(numHosts, thisHost, globPattern, cwd);
            break;
        default:
            console.error('No splitter provided');
    }
    return files;
}
exports.splitFiles = splitFiles;
