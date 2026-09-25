export const config = {
  siteUrl: process.env.NEXT_PUBLIC_SITE_URL || 'https://YOUR-DOMAIN.COM',
  siteName: process.env.NEXT_PUBLIC_SITE_NAME || 'Horizon',
  storeName: process.env.NEXT_PUBLIC_STORE_NAME || 'Horizon Store',
  apiUrl: process.env.HORIZON_API_URL || '',
  switchUrl: process.env.HORIZON_SWITCH_URL || '',
  orderApi: process.env.HORIZON_ORDER_API || '',
  issueApi: process.env.HORIZON_ISSUE_API || '',
  githubRepo: process.env.GITHUB_REPO_URL || 'https://github.com/beaconchain-horizon/horizon-core',
  githubRef: process.env.GITHUB_REF || '6269a737c7b422a0695baf948f6085446e344a52',
};
