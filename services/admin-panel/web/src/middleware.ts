import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const PUBLIC_PATHS = ['/login', '/_next', '/favicon.ico', '/api/'];

const ROLE_HIERARCHY: Record<string, number> = {
  super_admin: 4,
  product_manager: 3,
  order_manager: 2,
  viewer: 1,
};

const ROUTE_PERMISSIONS: Record<string, string[]> = {
  '/admin': ['super_admin', 'product_manager', 'order_manager', 'viewer'],
  '/admin/products': ['super_admin', 'product_manager'],
  '/admin/products/new': ['super_admin', 'product_manager'],
  '/admin/products/edit': ['super_admin', 'product_manager'],
  '/admin/orders': ['super_admin', 'order_manager'],
  '/admin/users': ['super_admin'],
  '/admin/analytics': ['super_admin', 'viewer'],
};

function isPublicPath(path: string): boolean {
  return PUBLIC_PATHS.some(p => path.startsWith(p));
}

function getRoutePermissions(path: string): string[] | null {
  for (const [route, roles] of Object.entries(ROUTE_PERMISSIONS)) {
    if (path === route || path.startsWith(route + '/')) {
      return roles;
    }
  }
  return null;
}

function hasToken(request: NextRequest): boolean {
  const cookie = request.cookies.get('access_token');
  return !!cookie?.value;
}

function getUserRole(request: NextRequest): string | null {
  return request.cookies.get('user_role')?.value ?? null;
}

function isAuthorized(path: string, role: string | null): boolean {
  if (!role) return false;
  const allowedRoles = getRoutePermissions(path);
  if (!allowedRoles) return true;
  return allowedRoles.includes(role);
}

export function middleware(request: NextRequest): NextResponse {
  const { pathname } = request.nextUrl;

  if (isPublicPath(pathname)) {
    return NextResponse.next();
  }

  if (!pathname.startsWith('/admin')) {
    return NextResponse.next();
  }

  if (!hasToken(request)) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  const role = getUserRole(request);

  if (!isAuthorized(pathname, role)) {
    const dashboardUrl = new URL('/admin', request.url);
    return NextResponse.redirect(dashboardUrl);
  }

  const response = NextResponse.next();

  response.headers.set('X-Content-Type-Options', 'nosniff');
  response.headers.set('X-Frame-Options', 'DENY');
  response.headers.set('X-XSS-Protection', '1; mode=block');
  response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
  response.headers.set(
    'Content-Security-Policy',
    "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https:; font-src 'self' data:;"
  );
  response.headers.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains');

  return response;
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
