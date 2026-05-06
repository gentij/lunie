import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import useBaseUrl from '@docusaurus/useBaseUrl';

import styles from './index.module.css';

const pathways = [
  {
    title: 'Get Running Fast',
    description:
      'Install Lunie, bootstrap the local stack, and run your first workflow in minutes.',
    href: '/docs/getting-started',
    label: 'Open Getting Started',
  },
  {
    title: 'Work From The CLI',
    description:
      'Create workflows, inspect runs, manage triggers, and work with secrets from the terminal.',
    href: '/docs/cli',
    label: 'Explore CLI Usage',
  },
  {
    title: 'Build On The Stack',
    description:
      'Understand the monorepo, local development flow, and runtime architecture behind Lunie.',
    href: '/docs/development',
    label: 'Read Development Docs',
  },
];

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  const heroArtwork = useBaseUrl('/img/logo.svg');

  return (
    <header className={styles.heroBanner}>
      <div className={clsx('container', styles.heroInner)}>
        <div className={styles.heroCopy}>
          <div className={styles.heroBadge}>Self-hosted workflow automation</div>
          <Heading as="h1" className={styles.heroTitle}>
            {siteConfig.title}
          </Heading>
          <p className={styles.heroSubtitle}>{siteConfig.tagline}</p>
          <p className={styles.heroLead}>
            Use Lunie when you want local control, API-first workflow definitions,
            and execution history you can actually inspect.
          </p>
          <div className={styles.heroActions}>
            <Link className={styles.primaryButton} to="/docs/getting-started">
              Install and Use Lunie
            </Link>
            <Link className={styles.secondaryButton} to="/docs/development">
              Contribute and Develop
            </Link>
          </div>
          <div className={styles.heroMeta}>
            <span>HTTP, transform, and condition steps</span>
            <span>CLI, TUI, server, and worker</span>
            <span>Runs and step history included</span>
          </div>
        </div>

        <div className={styles.heroArtWrap}>
          <div className={styles.heroArtFrame}>
            <img
              className={styles.heroArt}
              src={heroArtwork}
              alt="Lunie brand mark"
            />
          </div>
        </div>
      </div>
    </header>
  );
}

function HomepageMain() {
  return (
    <main className={styles.mainContent}>
      <section className={clsx('container', styles.pathwaysSection)}>
        <div className={styles.sectionHeading}>
          <Heading as="h2">Choose a path</Heading>
          <p>
            Start with the docs that match how you want to approach Lunie today.
          </p>
        </div>

        <div className={styles.cardGrid}>
          {pathways.map((path) => (
            <Link key={path.title} className={styles.pathCard} to={path.href}>
              <div>
                <Heading as="h3" className={styles.pathCardTitle}>
                  {path.title}
                </Heading>
                <p className={styles.pathCardBody}>{path.description}</p>
              </div>
              <span className={styles.pathCardAction}>{path.label}</span>
            </Link>
          ))}
        </div>
      </section>
    </main>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="Lunie product and developer documentation">
      <HomepageHeader />
      <HomepageMain />
    </Layout>
  );
}
