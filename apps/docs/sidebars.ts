import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'category',
      label: 'Start Here',
      items: ['getting-started', 'cli-usage', 'tui-guide', 'workflow-definitions'],
    },
    {
      type: 'category',
      label: 'Contributing',
      items: ['development-guide', 'architecture-overview', 'workflow-engine-mental-model'],
    },
    {
      type: 'category',
      label: 'References',
      items: [
        {
          type: 'link',
          label: 'CLI Guide (Repository)',
          href: 'https://github.com/gentij/taskforge/blob/main/apps/cli/README.md',
        },
      ],
    },
    {
      type: 'category',
      label: 'Archive',
      collapsed: true,
      items: ['initial-project-overview', 'queues-and-workers-plan'],
    },
  ],
};

export default sidebars;
