# Welcome

This repository contains the code required to run Brisk so you can host your own complete CI system. 

If you would prefer to use a hosted service instead of hosting your own CI system, check our hosted service at https://brisktest.com/.

Brisk is an extremely fast CI system, based around not rebuilding your environment on each test run. This allows us to really get the most from multiple workers. Instead of losing minutes rebuilding the environment on each run we instead can have the workers go straigth to work running tests. This dramatically shortens the total time a test run takes. With enough workers the speed of your longest test becomes the limit for how long your CI tests take. 

# Getting Started

The root of this repo contains a docker-compose.yml file which has a simple single worker deployment of Brisk. It is suitable for testing locally and can be used as a starting point for deploying to production. 

Brisk consists of several services that are designed to be run across hundreds to thousands of machines. The docker-compose.yml contains the simplest possible deployment, one of each of the services. 

To get started locally you can run 

```shell
docker compose up
```

# The CLI

In order to access Brisk from your project directory (where the source code you are trying to test lives) you'll need to use the CLI. You can download a prebuilt build from https://docs.brisktest.com/docs/installation or you can build the cli from this repo. 

In order to build the CLI you cd to 

```
cd core/brisk-cli
```

and execute ./build-debug to build a version for your system.

# Getting Started Using Brisk with your CI

Complete guides with information for setting up Brisk to work with your CI system and workflow are available at

https://docs.brisktest.com/

Examples include Github Actions, CircleCI, Bitbucket, AWS etc.

# Brisk High Level Architecture

Brisk consistes of a CLI program that talks to a an API and a dedicated supervisor (running in the cloud) which splits tests among many workers.
In production the CLI is the only componenet that runs locally (or in your CI Pipeline), everything else runs in the cloud. 

<?xml version="1.0" encoding="UTF-8"?>
<!-- Do not edit this file with editors other than draw.io -->
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="552px" height="421px" viewBox="-0.5 -0.5 552 421" content="&lt;mxfile host=&quot;app.diagrams.net&quot; modified=&quot;2024-05-22T03:59:48.023Z&quot; agent=&quot;Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36&quot; etag=&quot;lRvjlnWf3EhGpm3BsfNt&quot; version=&quot;24.4.4&quot; type=&quot;device&quot;&gt;&#10;  &lt;diagram name=&quot;Page-1&quot; id=&quot;nj3qzMIWaxASw7dmWYKP&quot;&gt;&#10;    &lt;mxGraphModel dx=&quot;954&quot; dy=&quot;496&quot; grid=&quot;1&quot; gridSize=&quot;10&quot; guides=&quot;1&quot; tooltips=&quot;1&quot; connect=&quot;1&quot; arrows=&quot;1&quot; fold=&quot;1&quot; page=&quot;1&quot; pageScale=&quot;1&quot; pageWidth=&quot;850&quot; pageHeight=&quot;1100&quot; math=&quot;0&quot; shadow=&quot;0&quot;&gt;&#10;      &lt;root&gt;&#10;        &lt;mxCell id=&quot;0&quot; /&gt;&#10;        &lt;mxCell id=&quot;1&quot; parent=&quot;0&quot; /&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-1&quot; value=&quot;&quot; style=&quot;rounded=1;whiteSpace=wrap;html=1;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;280&quot; y=&quot;100&quot; width=&quot;120&quot; height=&quot;60&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-2&quot; value=&quot;Actor&quot; style=&quot;shape=umlActor;verticalLabelPosition=bottom;verticalAlign=top;html=1;outlineConnect=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;190&quot; y=&quot;90&quot; width=&quot;30&quot; height=&quot;60&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-3&quot; value=&quot;&quot; style=&quot;endArrow=none;dashed=1;html=1;dashPattern=1 3;strokeWidth=2;rounded=0;&quot; edge=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry width=&quot;50&quot; height=&quot;50&quot; relative=&quot;1&quot; as=&quot;geometry&quot;&gt;&#10;            &lt;mxPoint x=&quot;130&quot; y=&quot;220&quot; as=&quot;sourcePoint&quot; /&gt;&#10;            &lt;mxPoint x=&quot;640&quot; y=&quot;220&quot; as=&quot;targetPoint&quot; /&gt;&#10;          &lt;/mxGeometry&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-5&quot; value=&quot;Brisk CLI&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;310&quot; y=&quot;115&quot; width=&quot;60&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-6&quot; value=&quot;&quot; style=&quot;rounded=1;whiteSpace=wrap;html=1;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;160&quot; y=&quot;280&quot; width=&quot;120&quot; height=&quot;60&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-7&quot; value=&quot;&quot; style=&quot;rounded=1;whiteSpace=wrap;html=1;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;310&quot; y=&quot;280&quot; width=&quot;120&quot; height=&quot;60&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-9&quot; value=&quot;&quot; style=&quot;rounded=0;whiteSpace=wrap;html=1;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;310&quot; y=&quot;450&quot; width=&quot;120&quot; height=&quot;60&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-10&quot; value=&quot;API&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;190&quot; y=&quot;295&quot; width=&quot;60&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-11&quot; value=&quot;Supervisor 1&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;340&quot; y=&quot;295&quot; width=&quot;70&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-12&quot; value=&quot;Worker 1&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;340&quot; y=&quot;465&quot; width=&quot;60&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-19&quot; value=&quot;&quot; style=&quot;endArrow=none;dashed=1;html=1;rounded=0;&quot; edge=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry width=&quot;50&quot; height=&quot;50&quot; relative=&quot;1&quot; as=&quot;geometry&quot;&gt;&#10;            &lt;mxPoint x=&quot;450&quot; y=&quot;480&quot; as=&quot;sourcePoint&quot; /&gt;&#10;            &lt;mxPoint x=&quot;550&quot; y=&quot;480&quot; as=&quot;targetPoint&quot; /&gt;&#10;          &lt;/mxGeometry&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-20&quot; value=&quot;1..n&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;460&quot; y=&quot;440&quot; width=&quot;60&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-21&quot; value=&quot;&amp;lt;span style=&amp;quot;color: rgb(0, 0, 0); font-family: Helvetica; font-size: 12px; font-style: normal; font-variant-ligatures: normal; font-variant-caps: normal; font-weight: 400; letter-spacing: normal; orphans: 2; text-align: center; text-indent: 0px; text-transform: none; widows: 2; word-spacing: 0px; -webkit-text-stroke-width: 0px; white-space: normal; background-color: rgb(251, 251, 251); text-decoration-thickness: initial; text-decoration-style: initial; text-decoration-color: initial; display: inline !important; float: none;&amp;quot;&amp;gt;Worker n&amp;lt;/span&amp;gt;&quot; style=&quot;rounded=0;whiteSpace=wrap;html=1;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;560&quot; y=&quot;450&quot; width=&quot;120&quot; height=&quot;60&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-22&quot; value=&quot;&amp;lt;span style=&amp;quot;color: rgb(0, 0, 0); font-family: Helvetica; font-size: 12px; font-style: normal; font-variant-ligatures: normal; font-variant-caps: normal; font-weight: 400; letter-spacing: normal; orphans: 2; text-align: center; text-indent: 0px; text-transform: none; widows: 2; word-spacing: 0px; -webkit-text-stroke-width: 0px; white-space: normal; background-color: rgb(251, 251, 251); text-decoration-thickness: initial; text-decoration-style: initial; text-decoration-color: initial; display: inline !important; float: none;&amp;quot;&amp;gt;Supervisor n&amp;lt;/span&amp;gt;&quot; style=&quot;rounded=1;whiteSpace=wrap;html=1;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;550&quot; y=&quot;280&quot; width=&quot;120&quot; height=&quot;60&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-23&quot; value=&quot;&quot; style=&quot;endArrow=none;dashed=1;html=1;rounded=0;&quot; edge=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry width=&quot;50&quot; height=&quot;50&quot; relative=&quot;1&quot; as=&quot;geometry&quot;&gt;&#10;            &lt;mxPoint x=&quot;440&quot; y=&quot;320&quot; as=&quot;sourcePoint&quot; /&gt;&#10;            &lt;mxPoint x=&quot;540&quot; y=&quot;320&quot; as=&quot;targetPoint&quot; /&gt;&#10;          &lt;/mxGeometry&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-24&quot; value=&quot;1..n&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;450&quot; y=&quot;280&quot; width=&quot;60&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-25&quot; value=&quot;&quot; style=&quot;endArrow=classic;html=1;rounded=0;&quot; edge=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry width=&quot;50&quot; height=&quot;50&quot; relative=&quot;1&quot; as=&quot;geometry&quot;&gt;&#10;            &lt;mxPoint x=&quot;370&quot; y=&quot;350&quot; as=&quot;sourcePoint&quot; /&gt;&#10;            &lt;mxPoint x=&quot;380&quot; y=&quot;440&quot; as=&quot;targetPoint&quot; /&gt;&#10;          &lt;/mxGeometry&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-26&quot; value=&quot;&quot; style=&quot;endArrow=classic;html=1;rounded=0;&quot; edge=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry width=&quot;50&quot; height=&quot;50&quot; relative=&quot;1&quot; as=&quot;geometry&quot;&gt;&#10;            &lt;mxPoint x=&quot;380&quot; y=&quot;350&quot; as=&quot;sourcePoint&quot; /&gt;&#10;            &lt;mxPoint x=&quot;590&quot; y=&quot;420&quot; as=&quot;targetPoint&quot; /&gt;&#10;          &lt;/mxGeometry&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-27&quot; value=&quot;&quot; style=&quot;endArrow=classic;html=1;rounded=0;&quot; edge=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry width=&quot;50&quot; height=&quot;50&quot; relative=&quot;1&quot; as=&quot;geometry&quot;&gt;&#10;            &lt;mxPoint x=&quot;330&quot; y=&quot;170&quot; as=&quot;sourcePoint&quot; /&gt;&#10;            &lt;mxPoint x=&quot;230&quot; y=&quot;270&quot; as=&quot;targetPoint&quot; /&gt;&#10;          &lt;/mxGeometry&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-28&quot; value=&quot;&quot; style=&quot;endArrow=classic;html=1;rounded=0;&quot; edge=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry width=&quot;50&quot; height=&quot;50&quot; relative=&quot;1&quot; as=&quot;geometry&quot;&gt;&#10;            &lt;mxPoint x=&quot;340&quot; y=&quot;170&quot; as=&quot;sourcePoint&quot; /&gt;&#10;            &lt;mxPoint x=&quot;360&quot; y=&quot;260&quot; as=&quot;targetPoint&quot; /&gt;&#10;          &lt;/mxGeometry&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-29&quot; value=&quot;Local or CI Pipeline&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;460&quot; y=&quot;160&quot; width=&quot;130&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;        &lt;mxCell id=&quot;PVIdiZpZ_IMOHBSRW37E-30&quot; value=&quot;Cloud/ On Prem Infra&quot; style=&quot;text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;&quot; vertex=&quot;1&quot; parent=&quot;1&quot;&gt;&#10;          &lt;mxGeometry x=&quot;470&quot; y=&quot;230&quot; width=&quot;120&quot; height=&quot;30&quot; as=&quot;geometry&quot; /&gt;&#10;        &lt;/mxCell&gt;&#10;      &lt;/root&gt;&#10;    &lt;/mxGraphModel&gt;&#10;  &lt;/diagram&gt;&#10;&lt;/mxfile&gt;&#10;"><defs/><g><g><rect x="151" y="10" width="120" height="60" rx="9" ry="9" fill="rgb(255, 255, 255)" stroke="rgb(0, 0, 0)" pointer-events="all"/></g><g><ellipse cx="76" cy="7.5" rx="7.5" ry="7.5" fill="rgb(255, 255, 255)" stroke="rgb(0, 0, 0)" pointer-events="all"/><path d="M 76 15 L 76 40 M 76 20 L 61 20 M 76 20 L 91 20 M 76 40 L 61 60 M 76 40 L 91 60" fill="none" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe flex-start; justify-content: unsafe center; width: 1px; height: 1px; padding-top: 67px; margin-left: 76px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: nowrap;">Actor</div></div></div></foreignObject><text x="76" y="79" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Actor</text></switch></g></g><g><path d="M 1 130 L 511 130" fill="none" stroke="rgb(0, 0, 0)" stroke-width="2" stroke-miterlimit="10" stroke-dasharray="2 6" pointer-events="stroke"/></g><g><rect x="181" y="25" width="60" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 58px; height: 1px; padding-top: 40px; margin-left: 182px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">Brisk CLI</div></div></div></foreignObject><text x="211" y="44" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Brisk CLI</text></switch></g></g><g><rect x="31" y="190" width="120" height="60" rx="9" ry="9" fill="rgb(255, 255, 255)" stroke="rgb(0, 0, 0)" pointer-events="all"/></g><g><rect x="181" y="190" width="120" height="60" rx="9" ry="9" fill="rgb(255, 255, 255)" stroke="rgb(0, 0, 0)" pointer-events="all"/></g><g><rect x="181" y="360" width="120" height="60" fill="rgb(255, 255, 255)" stroke="rgb(0, 0, 0)" pointer-events="all"/></g><g><rect x="61" y="205" width="60" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 58px; height: 1px; padding-top: 220px; margin-left: 62px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">API</div></div></div></foreignObject><text x="91" y="224" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">API</text></switch></g></g><g><rect x="211" y="205" width="70" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 68px; height: 1px; padding-top: 220px; margin-left: 212px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">Supervisor 1</div></div></div></foreignObject><text x="246" y="224" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Supervisor 1</text></switch></g></g><g><rect x="211" y="375" width="60" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 58px; height: 1px; padding-top: 390px; margin-left: 212px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">Worker 1</div></div></div></foreignObject><text x="241" y="394" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Worker 1</text></switch></g></g><g><path d="M 321 390 L 421 390" fill="none" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" stroke-dasharray="3 3" pointer-events="stroke"/></g><g><rect x="331" y="350" width="60" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 58px; height: 1px; padding-top: 365px; margin-left: 332px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">1..n</div></div></div></foreignObject><text x="361" y="369" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">1..n</text></switch></g></g><g><rect x="431" y="360" width="120" height="60" fill="rgb(255, 255, 255)" stroke="rgb(0, 0, 0)" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 118px; height: 1px; padding-top: 390px; margin-left: 432px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;"><span style="color: rgb(0, 0, 0); font-family: Helvetica; font-size: 12px; font-style: normal; font-variant-ligatures: normal; font-variant-caps: normal; font-weight: 400; letter-spacing: normal; orphans: 2; text-align: center; text-indent: 0px; text-transform: none; widows: 2; word-spacing: 0px; -webkit-text-stroke-width: 0px; white-space: normal; background-color: rgb(251, 251, 251); text-decoration-thickness: initial; text-decoration-style: initial; text-decoration-color: initial; display: inline !important; float: none;">Worker n</span></div></div></div></foreignObject><text x="491" y="394" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Worker n</text></switch></g></g><g><rect x="421" y="190" width="120" height="60" rx="9" ry="9" fill="rgb(255, 255, 255)" stroke="rgb(0, 0, 0)" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 118px; height: 1px; padding-top: 220px; margin-left: 422px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;"><span style="color: rgb(0, 0, 0); font-family: Helvetica; font-size: 12px; font-style: normal; font-variant-ligatures: normal; font-variant-caps: normal; font-weight: 400; letter-spacing: normal; orphans: 2; text-align: center; text-indent: 0px; text-transform: none; widows: 2; word-spacing: 0px; -webkit-text-stroke-width: 0px; white-space: normal; background-color: rgb(251, 251, 251); text-decoration-thickness: initial; text-decoration-style: initial; text-decoration-color: initial; display: inline !important; float: none;">Supervisor n</span></div></div></div></foreignObject><text x="481" y="224" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Supervisor n</text></switch></g></g><g><path d="M 311 230 L 411 230" fill="none" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" stroke-dasharray="3 3" pointer-events="stroke"/></g><g><rect x="321" y="190" width="60" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 58px; height: 1px; padding-top: 205px; margin-left: 322px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">1..n</div></div></div></foreignObject><text x="351" y="209" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">1..n</text></switch></g></g><g><path d="M 241 260 L 250.3 343.67" fill="none" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="stroke"/><path d="M 250.88 348.89 L 246.62 342.32 L 250.3 343.67 L 253.58 341.55 Z" fill="rgb(0, 0, 0)" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="all"/></g><g><path d="M 251 260 L 454.96 327.99" fill="none" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="stroke"/><path d="M 459.94 329.65 L 452.19 330.75 L 454.96 327.99 L 454.41 324.11 Z" fill="rgb(0, 0, 0)" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="all"/></g><g><path d="M 201 80 L 105.5 175.5" fill="none" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="stroke"/><path d="M 101.79 179.21 L 104.27 171.78 L 105.5 175.5 L 109.22 176.73 Z" fill="rgb(0, 0, 0)" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="all"/></g><g><path d="M 211 80 L 229.62 163.78" fill="none" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="stroke"/><path d="M 230.76 168.91 L 225.82 162.83 L 229.62 163.78 L 232.66 161.32 Z" fill="rgb(0, 0, 0)" stroke="rgb(0, 0, 0)" stroke-miterlimit="10" pointer-events="all"/></g><g><rect x="331" y="70" width="130" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 128px; height: 1px; padding-top: 85px; margin-left: 332px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">Local or CI Pipeline</div></div></div></foreignObject><text x="396" y="89" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Local or CI Pipeline</text></switch></g></g><g><rect x="341" y="140" width="120" height="30" fill="none" stroke="none" pointer-events="all"/></g><g><g transform="translate(-0.5 -0.5)"><switch><foreignObject pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility" style="overflow: visible; text-align: left;"><div xmlns="http://www.w3.org/1999/xhtml" style="display: flex; align-items: unsafe center; justify-content: unsafe center; width: 118px; height: 1px; padding-top: 155px; margin-left: 342px;"><div data-drawio-colors="color: rgb(0, 0, 0); " style="box-sizing: border-box; font-size: 0px; text-align: center;"><div style="display: inline-block; font-size: 12px; font-family: Helvetica; color: rgb(0, 0, 0); line-height: 1.2; pointer-events: all; white-space: normal; overflow-wrap: normal;">Cloud/ On Prem Infra</div></div></div></foreignObject><text x="401" y="159" fill="rgb(0, 0, 0)" font-family="Helvetica" font-size="12px" text-anchor="middle">Cloud/ On Prem Infra</text></switch></g></g></g><switch><g requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility"/><a transform="translate(0,-5)" xlink:href="https://www.drawio.com/doc/faq/svg-export-text-problems" target="_blank"><text text-anchor="middle" font-size="10px" x="50%" y="100%">Text is not SVG - cannot display</text></a></switch></svg>

# DEV mode

There are a few reasons why you would like to run brisk with different deployment requirements. 

First off you could be contributing code to the project.

Second, you could be trying out the project to see if it is suitable for your use case and 

Finally you could be running an instance of the project in production, perhaps in a private cloud or on some other infra.

The docker-compose deployment (which relies on DEV mode) is suitable for the first two objectives but is UNSUITABLE for deployment to production. 

To make the local setup easier many important security features are disabled in DEV mode (such as tls certs, isolation of different parts of the system, many safety mechanisms and fallbacks). Please do not run the system in DEV mode in production. 

In order to safely run Brisk in production you need to turn DEV mode off and implement relevant security measures. At a minimum configure TLS with certificates. 

