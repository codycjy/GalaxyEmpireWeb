'use client';

import { SessionProvider } from "next-auth/react";
import { UserProvider } from './user-provider';

export default function Providers({ 
  children,
  session 
}: { 
  children: React.ReactNode;
  session: any;
}) {
  return (
    <SessionProvider session={session}>
      <UserProvider>
        {children}
      </UserProvider>
    </SessionProvider>
  );
}