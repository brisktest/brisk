"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.config = void 0;
const findProjectRoot_1 = require("./findProjectRoot");
let config;
exports.config = config;
try {
    exports.config = config = require(findProjectRoot_1.findProjectRoot(process.cwd.toString()) +
        '/more-faster.json');
}
catch (e) {
    console.log(e.message);
    exports.config = config = {};
}
console.log(config);
