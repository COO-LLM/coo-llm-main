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
      label: '🚀 Getting Started',
      collapsed: false,
      items: [
        'Intro/Overview',
        'Getting-Started/Installation',
        'Getting-Started/Quick-Start',
        'Getting-Started/First-Steps',
        'Getting-Started/Whats-Next',
        'Guides/Deployment',
      ],
    },
    {
      type: 'category',
      label: '👤 User Guide',
      collapsed: false,
      items: [
        'User-Guide/API-Usage',
        'User-Guide/Practical-Usage',
        'Guides/Providers',
        'User-Guide/Examples',
      ],
    },
    {
      type: 'category',
      label: '⚙️ Administrator Guide',
      collapsed: false,
      items: [
        'Administrator-Guide/Overview',
        'Guides/Configuration',
        'Administrator-Guide/Monitoring',
        'Administrator-Guide/Troubleshooting',
        'Administrator-Guide/Web-UI',
      ],
    },
    {
      type: 'category',
      label: '🔧 Developer Guide',
      collapsed: false,
      items: [
        'Intro/Architecture',
        {
          type: 'category',
          label: 'Providers',
          items: [
            'Developer-Guide/Providers/OpenAI',
            'Developer-Guide/Providers/Gemini',
            'Developer-Guide/Providers/Claude',
            'Developer-Guide/Providers/Grok',
            'Developer-Guide/Providers/Together',
            'Developer-Guide/Providers/OpenRouter',
            'Developer-Guide/Providers/Mistral',
            'Developer-Guide/Providers/Cohere',
            'Developer-Guide/Providers/HuggingFace',
            'Developer-Guide/Providers/Replicate',
            'Developer-Guide/Providers/Voyage',
            'Developer-Guide/Providers/Fireworks',
          ],
        },
        {
          type: 'category',
          label: 'API Reference',
          items: [
            'Reference/Admin-API',
            'Reference/LLM-API',
          ],
        },
        'Reference/Balancer',
        {
          type: 'category',
          label: 'Storage',
          items: [
            'Reference/Storage',
            'Reference/Storage/Redis',
            'Reference/Storage/MongoDB',
            'Reference/Storage/DynamoDB',
            'Reference/Storage/InfluxDB',
            'Reference/Storage/HTTP',
            'Reference/Storage/File',
            'Reference/Storage/SQL',
          ],
        },
        {
          type: 'category',
          label: 'Logging',
          items: [
            'Reference/Logging',
            'Reference/Logging/Request-Logging',
            'Reference/Logging/Usage-Logging',
            'Reference/Logging/Prometheus',
          ],
        },
        'Developer-Guide/Testing',
        'Contributing/Guidelines',
        'Contributing/Changelog',
      ],
    },
    {
      type: 'category',
      label: '📚 Reference',
      collapsed: false,
      items: [
        'Reference/Config-Schema',
        'Reference/Error-Codes',
        'Reference/Glossary',
      ],
    },
  ],
};

module.exports = sidebars;