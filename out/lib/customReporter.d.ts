export default class JestProgressBarReporter {
    _globalConfig: any;
    _options: any;
    _numTotalTestSuites: any;
    _bar: any;
    constructor(globalConfig: any, options: any);
    onRunStart(test: {
        numTotalTestSuites: any;
    }): void;
    onTestStart(): void;
    onRunComplete(test: any, results: {
        numFailedTests: any;
        numPassedTests: any;
        numPendingTests: any;
        testResults: any;
        numTotalTests: any;
        startTime: any;
        snapshot: any;
    }): void;
    onTestResult(): void;
    _getStatus(status: string): any;
}
