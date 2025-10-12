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
        'Intro/Overview',
        'Guides/Configuration',
        'Guides/Deployment',
        'Guides/Providers',
      ],
    },
    {
      type: 'category',
      label: '🔧 Developer Guide',
      collapsed: false,
      items: [
        'Intro/Architecture',
        'Reference/API',
        'Reference/Balancer',
        'Reference/Storage',
        'Reference/Logging',
        'Contributing/Guidelines',
        'Contributing/Changelog',
      ],
    },
  ],
};

module.exports = sidebars;