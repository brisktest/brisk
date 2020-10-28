#!/usr/bin/env node
import { config, getHostnames } from './config'
import { startSyncForHosts } from './lib/synFs'

const hosts = getHostnames()
const localPath = config.projectRoot
const remotePath = config.remoteWorkspace
startSyncForHosts(hosts, localPath, remotePath)
