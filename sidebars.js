/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-nocheck

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    {
      type: 'category',
      label: '🚀 User Guide',
      collapsed: false,
      items: [
        'intro/Overview',
        'guides/Configuration',
        'guides/Deployment',
        'guides/Providers',
      ],
    },
    {
      type: 'category',
      label: '🔧 Developer Guide',
      collapsed: false,
      items: [
        'intro/Architecture',
        'reference/API',
        'reference/Balancer',
        'reference/Storage',
        'reference/Logging',
        'contributing/Guidelines',
        'contributing/Changelog',
      ],
    },
  ],
};

module.exports = sidebars;