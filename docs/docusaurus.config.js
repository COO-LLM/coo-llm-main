const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').DocusaurusConfig} */
const config = {
  title: 'COO-LLM',
  tagline: 'Intelligent LLM API Load Balancing',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://coo-llm.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/docs',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
   organizationName: 'coo-llm', // Usually your GitHub org/user name.
   projectName: 'docs', // Usually your repo name.

  onBrokenLinks: 'ignore',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },
  markdown: {
    mermaid: true,
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },
  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          path: 'content',
          sidebarPath: './sidebars.js',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/coo-llm/coo-llm-main/tree/main/docs/content/',
          remarkPlugins: [require('remark-mermaid')],
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      })
    ],
  ],

  themes: ['@docusaurus/theme-mermaid'],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      image: 'img/coo-llm-social-card.jpg',
      navbar: {
        title: 'COO-LLM',
        logo: {
          alt: 'COO-LLM Logo',
          src: '/img/logo.png',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'tutorialSidebar',
            position: 'left',
            label: 'Docs',
          },
          {
            href: 'https://github.com/COO-LLM/coo-llm-main',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Getting Started',
                to: '/docs/intro/overview',
              },
              {
                label: 'Configuration',
                to: '/docs/guides/configuration',
              },
              {
                label: 'API Reference',
                to: '/docs/reference/api',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/COO-LLM/coo-llm-main',
              },
              {
                label: 'Issues',
                href: 'https://github.com/COO-LLM/coo-llm-main/issues',
              },
              {
                label: 'Discussions',
                href: 'https://github.com/COO-LLM/coo-llm-main/discussions',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'Contributing',
                to: '/docs/contributing/guidelines',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} COO-LLM.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      mermaid: {
        options: {
          theme: 'default',
        },
      },

    }),
};

module.exports = config;