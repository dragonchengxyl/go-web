import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const PROTECTED = [
  '/feed',
  '/posts/create',
  '/messages',
  '/notifications',
  '/profile',
  '/settings',
  '/creator',
];

export function middleware(request: NextRequest) {
  const isAuth = request.cookies.has('_auth');
  const { pathname } = request.nextUrl;

  // Authenticated users away from auth pages
  if (isAuth && (pathname === '/login' || pathname === '/register')) {
    return NextResponse.redirect(new URL('/feed', request.url));
  }

  // Unauthenticated users away from protected pages
  if (!isAuth && PROTECTED.some(p => pathname === p || pathname.startsWith(p + '/'))) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('from', pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/login',
    '/register',
    '/feed',
    '/posts/create',
    '/messages/:path*',
    '/notifications',
    '/profile',
    '/settings',
    '/creator',
  ],
};
