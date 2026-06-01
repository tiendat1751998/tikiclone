import { redirect } from 'next/navigation';
import { cookies } from 'next/headers';
import { LoginForm } from './LoginForm';

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ redirect?: string; error?: string }>;
}) {
  const cookieStore = await cookies();
  const token = cookieStore.get('access_token');

  if (token) {
    const params = await searchParams;
    redirect(params.redirect || '/admin');
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 via-background to-primary-100 dark:from-background dark:via-background dark:to-primary-950 p-4">
      <div className="w-full max-w-md">
        <div className="rounded-2xl border border-border bg-card shadow-xl p-8">
          <div className="text-center mb-8">
            <div className="w-16 h-16 rounded-2xl bg-primary-500 flex items-center justify-center mx-auto mb-4">
              <span className="text-white font-bold text-2xl">T</span>
            </div>
            <h1 className="text-2xl font-bold text-foreground">Tiki Admin</h1>
            <p className="text-sm text-muted-foreground mt-2">
              Sign in to access the admin dashboard
            </p>
          </div>

          <LoginForm />
        </div>

        <p className="text-center text-xs text-muted-foreground mt-6">
          Protected area. Only authorized personnel may access.
        </p>
      </div>
    </div>
  );
}
