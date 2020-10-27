"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.config = exports.getHostnames = void 0;
const findProjectRoot_1 = require("./findProjectRoot");
let config;
exports.config = config;
const projectRoot = findProjectRoot_1.findProjectRoot(process.cwd.toString());
try {
    exports.config = config = require(projectRoot +
        '/more-faster.json');
}
catch (e) {
    console.log(e.message);
    exports.config = config = {};
}
config.projectRoot = projectRoot;
console.log(config);
function getHostnames() {
    return Object.keys(config.hosts);
}
exports.getHostnames = getHostnames;
