// import { searchParamsCache } from '@/lib/searchparams';
import { SearchParams } from 'nuqs/parsers';
import React from 'react';
import AccountListingPage from './_components/account-listing-page';

type pageProps = {
  searchParams: SearchParams;
};

export const metadata = {
  title: 'Dashboard : Accounts'
};

export default async function Page({ searchParams }: pageProps) {
  // TODO: Allow nested RSCs to access the search params (in a type-safe way)
  // searchParamsCache.parse(searchParams);

  return <AccountListingPage />;
}
