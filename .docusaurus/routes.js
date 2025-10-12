import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/truckllm/__docusaurus/debug',
    component: ComponentCreator('/truckllm/__docusaurus/debug', '04a'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/config',
    component: ComponentCreator('/truckllm/__docusaurus/debug/config', '21b'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/content',
    component: ComponentCreator('/truckllm/__docusaurus/debug/content', '8d5'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/globalData',
    component: ComponentCreator('/truckllm/__docusaurus/debug/globalData', '96e'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/metadata',
    component: ComponentCreator('/truckllm/__docusaurus/debug/metadata', '35b'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/registry',
    component: ComponentCreator('/truckllm/__docusaurus/debug/registry', 'dc2'),
    exact: true
  },
  {
    path: '/truckllm/__docusaurus/debug/routes',
    component: ComponentCreator('/truckllm/__docusaurus/debug/routes', '99f'),
    exact: true
  },
  {
    path: '/truckllm/docs',
    component: ComponentCreator('/truckllm/docs', '180'),
    routes: [
      {
        path: '/truckllm/docs/',
        component: ComponentCreator('/truckllm/docs/', '27b'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/contributing',
        component: ComponentCreator('/truckllm/docs/contributing', '4e9'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/contributing/changelog',
        component: ComponentCreator('/truckllm/docs/contributing/changelog', 'ca8'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/contributing/guidelines',
        component: ComponentCreator('/truckllm/docs/contributing/guidelines', 'a79'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/guides',
        component: ComponentCreator('/truckllm/docs/guides', 'f07'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/guides/configuration',
        component: ComponentCreator('/truckllm/docs/guides/configuration', '855'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/guides/deployment',
        component: ComponentCreator('/truckllm/docs/guides/deployment', 'cf5'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/guides/providers',
        component: ComponentCreator('/truckllm/docs/guides/providers', '658'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/intro',
        component: ComponentCreator('/truckllm/docs/intro', 'be5'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/intro/architecture',
        component: ComponentCreator('/truckllm/docs/intro/architecture', '6e9'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/intro/overview',
        component: ComponentCreator('/truckllm/docs/intro/overview', '6e8'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/reference',
        component: ComponentCreator('/truckllm/docs/reference', '646'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/reference/api',
        component: ComponentCreator('/truckllm/docs/reference/api', 'd33'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/reference/balancer',
        component: ComponentCreator('/truckllm/docs/reference/balancer', 'df5'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/reference/logging',
        component: ComponentCreator('/truckllm/docs/reference/logging', '801'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/truckllm/docs/reference/storage',
        component: ComponentCreator('/truckllm/docs/reference/storage', '65b'),
        exact: true,
        sidebar: "tutorialSidebar"
      }
    ]
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
