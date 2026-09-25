import './globals.css';
import { SiteNav } from '../components/SiteNav';

export const metadata = {
  title: 'Horizon — Secure Infrastructure',
  description: 'Horizon Store and Horizon Angel private SOC',
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="fa" dir="rtl"><body><SiteNav />{children}<footer className="footer"><div className="container"><div className="split"><span>Horizon · Secure Infrastructure</span><span>Domain & API are configurable via environment variables.</span></div></div></footer></body></html>;
}
