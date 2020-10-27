"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.prettyPrintOutput = void 0;
function prettyPrintOutput(jsonArray) {
    console.log(JSON.parse(jsonArray.join(' ')));
}
exports.prettyPrintOutput = prettyPrintOutput;
