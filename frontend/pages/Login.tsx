import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Loader2, ArrowRight, Sparkles, Mail, Lock } from 'lucide-react';

export const Login: React.FC = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsSubmitting(true);
        try {
            await login({ email, password });
            navigate('/');
        } catch (err) {
            setError('Credenciais invalidas. Tente novamente.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-background py-8 px-4">
            <div className="w-full max-w-md">
                {/* Logo */}
                <div className="text-center mb-8">
                    <div className="inline-flex p-4 bg-gradient-primary rounded-2xl shadow-card mb-4">
                        <Sparkles className="w-8 h-8 text-white" />
                    </div>
                    <h1 className="text-3xl font-bold text-foreground">
                        MyPila<span className="text-primary-500">Pro</span>
                    </h1>
                </div>

                {/* Login Card */}
                <div className="card p-8 animate-fadeIn">
                    <div className="text-center mb-8">
                        <h2 className="text-2xl font-bold text-foreground mb-2">Bem-vindo de volta</h2>
                        <p className="text-muted">Entre para gerenciar suas financas</p>
                    </div>

                    {error && (
                        <div className="bg-destructive-light border border-destructive/30 text-destructive px-4 py-3 rounded-xl mb-6 text-sm flex items-center gap-2">
                            <span className="w-5 h-5 rounded-full bg-destructive/20 flex items-center justify-center text-xs">!</span>
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-5">
                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2 ml-1">Email</label>
                            <div className="relative">
                                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted" />
                                <input
                                    type="email"
                                    required
                                    className="input pl-12"
                                    placeholder="seu@email.com"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                />
                            </div>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2 ml-1">Senha</label>
                            <div className="relative">
                                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted" />
                                <input
                                    type="password"
                                    required
                                    className="input pl-12"
                                    placeholder="********"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                />
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={isSubmitting}
                            className="w-full btn-primary flex items-center justify-center gap-2"
                        >
                            {isSubmitting ? (
                                <Loader2 className="w-5 h-5 animate-spin" />
                            ) : (
                                <>
                                    Entrar
                                    <ArrowRight className="w-5 h-5" />
                                </>
                            )}
                        </button>
                    </form>

                    <div className="mt-8 text-center">
                        <p className="text-muted text-sm">
                            Nao tem uma conta?{' '}
                            <Link to="/register" className="text-primary-600 hover:text-primary-700 font-medium transition-colors">
                                Cadastre-se gratis
                            </Link>
                        </p>
                    </div>
                </div>

                {/* Footer */}
                <p className="text-center text-muted text-xs mt-8">
                    Gestao financeira inteligente
                </p>
            </div>
        </div>
    );
};
