import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import { z } from 'zod';
import { useAuth } from './AuthContext';

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(1)
});

type LoginForm = z.infer<typeof schema>;

export function LoginPage() {
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState('');
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm<LoginForm>({
    resolver: zodResolver(schema),
    defaultValues: { email: 'admin@example.com', password: 'admin123' }
  });

  if (user) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <main className="loginPage">
      <section className="loginPanel">
        <div>
          <span className="eyebrow">Asset Management System</span>
          <h1>Sign in</h1>
        </div>
        <form
          onSubmit={handleSubmit(async (values) => {
            setError('');
            try {
              await login(values.email, values.password);
              const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname;
              navigate(from ?? '/dashboard', { replace: true });
            } catch (err) {
              setError(err instanceof Error ? err.message : 'Login failed');
            }
          })}
        >
          <label>
            Email
            <input autoComplete="email" {...register('email')} />
            {errors.email && <span className="fieldError">{errors.email.message}</span>}
          </label>
          <label>
            Password
            <input type="password" autoComplete="current-password" {...register('password')} />
            {errors.password && <span className="fieldError">{errors.password.message}</span>}
          </label>
          {error && <div className="alert">{error}</div>}
          <button className="primaryButton" disabled={isSubmitting} type="submit">
            {isSubmitting ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
        <p className="muted">Demo credentials: admin@example.com / admin123 or user@example.com / user123</p>
      </section>
    </main>
  );
}
