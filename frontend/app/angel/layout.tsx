import {requireAngel} from '../../lib/angel-auth';
export default async function AngelLayout({children}:{children:React.ReactNode}){await requireAngel();return <>{children}</>}
