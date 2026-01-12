import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Loader2, ArrowLeft, Mail, Sparkles, CheckCircle } from 'lucide-react';
import { api } from '../services/api';

export const ForgotPassword: React.FC = () => {
    const [email, setEmail] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsSubmitting(true);

        try {
            await api.forgotPassword(email);
            setSuccess(true);
        } catch (err: any) {
            // Always show success to prevent email enumeration
            setSuccess(true);
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
                            <h2 className="text-2xl font-bold text-foreground mb-2">Email Enviado!</h2>
                            <p className="text-muted mb-6">
                                Se o email <strong>{email}</strong> estiver cadastrado, voce recebera um link para redefinir sua senha.
                            </p>
                            <p className="text-muted text-sm mb-6">
                                Verifique sua caixa de entrada e a pasta de spam.
                            </p>
                            <Link
                                to="/login"
                                className="btn-primary inline-flex items-center gap-2"
                            >
                                <ArrowLeft className="w-5 h-5" />
                                Voltar ao Login
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
                        <h2 className="text-2xl font-bold text-foreground mb-2">Esqueceu a senha?</h2>
                        <p className="text-muted">Digite seu email para recuperar sua conta</p>
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

                        <button
                            type="submit"
                            disabled={isSubmitting}
                            className="w-full btn-primary flex items-center justify-center gap-2"
                        >
                            {isSubmitting ? (
                                <Loader2 className="w-5 h-5 animate-spin" />
                            ) : (
                                'Enviar Link de Recuperacao'
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
