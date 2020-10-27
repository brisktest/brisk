#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const config_1 = require("./config");
const synFs_1 = require("./lib/synFs");
const hosts = config_1.getHostnames();
const localPath = config_1.config.projectRoot;
const remotePath = config_1.config.remoteWorkspace;
synFs_1.startSyncForHosts(hosts, localPath, remotePath);
