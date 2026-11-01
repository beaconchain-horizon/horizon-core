import { requireAngel } from '../../lib/angel-auth';
import AngelPageClient from './page-client';

export default async function AngelPage() {
  await requireAngel();
  return <AngelPageClient />;
}
