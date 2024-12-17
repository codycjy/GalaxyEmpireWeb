// Protecting routes with next-auth
// https://next-auth.js.org/configuration/nextjs#middleware
// https://nextjs.org/docs/app/building-your-application/routing/middleware

import { auth } from './auth.config';

export { auth as middleware };

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)']
};
