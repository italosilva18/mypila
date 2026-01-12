import React, { useState, useEffect } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { Loader2, ArrowLeft, Lock, Sparkles, CheckCircle, XCircle } from 'lucide-react';
import { api } from '../services/api';

export const ResetPassword: React.FC = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const token = searchParams.get('token');

    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);

    useEffect(() => {
        if (!token) {
            setError('Token de recuperacao invalido ou ausente.');
        }
    }, [token]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (newPassword !== confirmPassword) {
            setError('As senhas nao coincidem.');
            return;
        }

        if (newPassword.length < 6) {
            setError('A senha deve ter pelo menos 6 caracteres.');
            return;
        }

        setIsSubmitting(true);

        try {
            await api.resetPassword(token!, newPassword);
            setSuccess(true);
        } catch (err: any) {
            const message = err.response?.data?.error || 'Erro ao redefinir senha. Tente novamente.';
            setError(message);
        } finally {
            setIsSubmitting(false);
        }
    };

    if (success) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-background py-8 px-4">
                <div className="w-full max-w-md">
                    <div className="text-center mb-8">
                        <div className="inline-flex p-4 bg-gradient-primary rounded-2xl shadow-card mb-4">
                            <Sparkles className="w-8 h-8 text-white" />
                        </div>
                        <h1 className="text-3xl font-bold text-foreground">
                            MyPila<span className="text-primary-500">Pro</span>
                        </h1>
                    </div>

                    <div className="card p-8 animate-fadeIn">
                        <div className="text-center">
                            <div className="inline-flex p-4 bg-green-100 rounded-full mb-4">
                                <CheckCircle className="w-8 h-8 text-green-600" />
                            </div>
                            <h2 className="text-2xl font-bold text-foreground mb-2">Senha Redefinida!</h2>
                            <p className="text-muted mb-6">
                                Sua senha foi alterada com sucesso. Agora voce pode fazer login com sua nova senha.
                            </p>
                            <button
                                onClick={() => navigate('/login')}
                                className="btn-primary inline-flex items-center gap-2"
                            >
                                Ir para Login
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    if (!token) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-background py-8 px-4">
                <div className="w-full max-w-md">
                    <div className="text-center mb-8">
                        <div className="inline-flex p-4 bg-gradient-primary rounded-2xl shadow-card mb-4">
                            <Sparkles className="w-8 h-8 text-white" />
                        </div>
                        <h1 className="text-3xl font-bold text-foreground">
                            MyPila<span className="text-primary-500">Pro</span>
                        </h1>
                    </div>

                    <div className="card p-8 animate-fadeIn">
                        <div className="text-center">
                            <div className="inline-flex p-4 bg-red-100 rounded-full mb-4">
                                <XCircle className="w-8 h-8 text-red-600" />
                            </div>
                            <h2 className="text-2xl font-bold text-foreground mb-2">Link Invalido</h2>
                            <p className="text-muted mb-6">
                                O link de recuperacao de senha e invalido ou expirou.
                            </p>
                            <Link
                                to="/forgot-password"
                                className="btn-primary inline-flex items-center gap-2"
                            >
                                Solicitar Novo Link
                            </Link>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-background py-8 px-4">
            <div className="w-full max-w-md">
                <div className="text-center mb-8">
                    <div className="inline-flex p-4 bg-gradient-primary rounded-2xl shadow-card mb-4">
                        <Sparkles className="w-8 h-8 text-white" />
                    </div>
                    <h1 className="text-3xl font-bold text-foreground">
                        MyPila<span className="text-primary-500">Pro</span>
                    </h1>
                </div>

                <div className="card p-8 animate-fadeIn">
                    <div className="text-center mb-8">
                        <h2 className="text-2xl font-bold text-foreground mb-2">Nova Senha</h2>
                        <p className="text-muted">Digite sua nova senha</p>
                    </div>

                    {error && (
                        <div className="bg-destructive-light border border-destructive/30 text-destructive px-4 py-3 rounded-xl mb-6 text-sm flex items-center gap-2">
                            <span className="w-5 h-5 rounded-full bg-destructive/20 flex items-center justify-center text-xs">!</span>
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-5">
                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2 ml-1">Nova Senha</label>
                            <div className="relative">
                                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted" />
                                <input
                                    type="password"
                                    required
                                    minLength={6}
                                    className="input pl-12"
                                    placeholder="********"
                                    value={newPassword}
                                    onChange={(e) => setNewPassword(e.target.value)}
                                />
                            </div>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2 ml-1">Confirmar Senha</label>
                            <div className="relative">
                                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted" />
                                <input
                                    type="password"
                                    required
                                    minLength={6}
                                    className="input pl-12"
                                    placeholder="********"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
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
                                'Redefinir Senha'
                            )}
                        </button>
                    </form>

                    <div className="mt-8 text-center">
                        <Link
                            to="/login"
                            className="text-primary-600 hover:text-primary-700 font-medium transition-colors inline-flex items-center gap-2"
                        >
                            <ArrowLeft className="w-4 h-4" />
                            Voltar ao Login
                        </Link>
                    </div>
                </div>

                <p className="text-center text-muted text-xs mt-8">
                    Gestao financeira inteligente
                </p>
            </div>
        </div>
    );
};
