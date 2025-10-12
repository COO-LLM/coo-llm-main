import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/truckllm/__docusaurus/debug',
    component: ComponentCreator('/truckllm/__docusaurus/debug', '70d'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/config',
    component: ComponentCreator('/truckllm/__docusaurus/debug/config', 'ec8'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/content',
    component: ComponentCreator('/truckllm/__docusaurus/debug/content', '8d7'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/globalData',
    component: ComponentCreator('/truckllm/__docusaurus/debug/globalData', 'f75'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/metadata',
    component: ComponentCreator('/truckllm/__docusaurus/debug/metadata', 'aef'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/registry',
    component: ComponentCreator('/truckllm/__docusaurus/debug/registry', 'dfa'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/routes',
    component: ComponentCreator('/truckllm/__docusaurus/debug/routes', 'ca0'),
    exact: true
  },
  {
    path: '/truckllm/docs',
    component: ComponentCreator('/truckllm/docs', '9b5'),
    routes: [
      {
        path: '/truckllm/docs',
        component: ComponentCreator('/truckllm/docs', '0e8'),
        routes: [
          {
            path: '/truckllm/docs',
            component: ComponentCreator('/truckllm/docs', 'd13'),
            routes: [
              {
                path: '/truckllm/docs/',
                component: ComponentCreator('/truckllm/docs/', '700'),
                exact: true
              },
              {
                path: '/truckllm/docs/Contributing',
                component: ComponentCreator('/truckllm/docs/Contributing', '18f'),
                exact: true
              },
              {
                path: '/truckllm/docs/Contributing/Changelog',
                component: ComponentCreator('/truckllm/docs/Contributing/Changelog', 'ad2'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Contributing/Guidelines',
                component: ComponentCreator('/truckllm/docs/Contributing/Guidelines', '5a2'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Guides',
                component: ComponentCreator('/truckllm/docs/Guides', '29e'),
                exact: true
              },
              {
                path: '/truckllm/docs/Guides/Configuration',
                component: ComponentCreator('/truckllm/docs/Guides/Configuration', '38b'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Guides/Deployment',
                component: ComponentCreator('/truckllm/docs/Guides/Deployment', '836'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Guides/Providers',
                component: ComponentCreator('/truckllm/docs/Guides/Providers', '454'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Intro',
                component: ComponentCreator('/truckllm/docs/Intro', '99a'),
                exact: true
              },
              {
                path: '/truckllm/docs/Intro/Architecture',
                component: ComponentCreator('/truckllm/docs/Intro/Architecture', 'cde'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Intro/Overview',
                component: ComponentCreator('/truckllm/docs/Intro/Overview', '34e'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Reference',
                component: ComponentCreator('/truckllm/docs/Reference', '7da'),
                exact: true
              },
              {
                path: '/truckllm/docs/Reference/API',
                component: ComponentCreator('/truckllm/docs/Reference/API', '21b'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Reference/Balancer',
                component: ComponentCreator('/truckllm/docs/Reference/Balancer', '520'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Reference/Logging',
                component: ComponentCreator('/truckllm/docs/Reference/Logging', '366'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/truckllm/docs/Reference/Storage',
                component: ComponentCreator('/truckllm/docs/Reference/Storage', '42d'),
                exact: true,
                sidebar: "tutorialSidebar"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
