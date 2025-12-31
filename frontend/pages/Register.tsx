import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Loader2, ArrowRight } from 'lucide-react';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMinLength, validateMaxLength, combineValidations } from '../utils/validation';
import { ErrorMessage } from '../components/ErrorMessage';

export const Register: React.FC = () => {
    const [name, setName] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const { register } = useAuth();
    const navigate = useNavigate();
    const { validateFields, getError, hasError } = useFormValidation();

    const validateForm = (): boolean => {
        return validateFields({
            name: () => combineValidations(
                validateRequired(name, 'Nome'),
                validateMaxLength(name, 100, 'Nome')
            ),
            email: () => validateRequired(email, 'Email'),
            password: () => combineValidations(
                validateRequired(password, 'Senha'),
                validateMinLength(password, 6, 'Senha')
            )
        });
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (!validateForm()) return;

        setIsSubmitting(true);
        try {
            await register({ name, email, password });
            navigate('/');
        } catch (err) {
            setError('Erro ao criar conta. Tente novamente.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-paper relative overflow-hidden">
            {/* Background Effects - Vintage style */}
            <div className="absolute top-0 left-0 w-full h-full overflow-hidden z-0">
                <div className="absolute top-[10%] -left-[10%] w-[60%] h-[60%] rounded-full bg-stone-300/30 blur-[120px]" />
                <div className="absolute bottom-[10%] -right-[10%] w-[60%] h-[60%] rounded-full bg-stone-400/20 blur-[120px]" />
            </div>

            <div className="w-full max-w-md px-4 md:p-8 z-10">
                <div className="bg-white/80 backdrop-blur-xl border border-stone-200 rounded-2xl md:rounded-3xl shadow-card p-6 md:p-8 transform transition-all hover:shadow-lg">
                    <div className="text-center mb-6 md:mb-8">
                        <h2 className="text-2xl md:text-3xl font-bold text-stone-900 mb-1 md:mb-2 tracking-tight">Criar Conta</h2>
                        <p className="text-stone-500 text-sm md:text-base">Comece a controlar suas finanças hoje</p>
                    </div>

                    {error && (
                        <div className="bg-red-50 border border-red-200 text-red-700 px-3 md:px-4 py-2.5 md:py-3 rounded-lg md:rounded-xl mb-4 md:mb-6 text-xs md:text-sm flex items-center gap-2">
                            <span>!</span> {error}
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-4 md:space-y-5">
                        <div>
                            <label className="block text-xs md:text-sm font-medium text-stone-700 mb-1.5 md:mb-2 ml-1">Nome Completo</label>
                            <input
                                type="text"
                                className={`w-full px-3 md:px-4 py-3 rounded-lg md:rounded-xl bg-stone-50 border text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:border-transparent transition-all ${
                                    hasError('name')
                                        ? 'border-red-500 focus:ring-red-400'
                                        : 'border-stone-200 focus:ring-stone-400'
                                }`}
                                placeholder="Seu nome"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                            />
                            <ErrorMessage error={getError('name')} />
                        </div>

                        <div>
                            <label className="block text-xs md:text-sm font-medium text-stone-700 mb-1.5 md:mb-2 ml-1">Email</label>
                            <input
                                type="email"
                                className={`w-full px-3 md:px-4 py-3 rounded-lg md:rounded-xl bg-stone-50 border text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:border-transparent transition-all ${
                                    hasError('email')
                                        ? 'border-red-500 focus:ring-red-400'
                                        : 'border-stone-200 focus:ring-stone-400'
                                }`}
                                placeholder="seu@email.com"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                            />
                            <ErrorMessage error={getError('email')} />
                        </div>

                        <div>
                            <label className="block text-xs md:text-sm font-medium text-stone-700 mb-1.5 md:mb-2 ml-1">Senha</label>
                            <input
                                type="password"
                                className={`w-full px-3 md:px-4 py-3 rounded-lg md:rounded-xl bg-stone-50 border text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:border-transparent transition-all ${
                                    hasError('password')
                                        ? 'border-red-500 focus:ring-red-400'
                                        : 'border-stone-200 focus:ring-stone-400'
                                }`}
                                placeholder="********"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                            />
                            <ErrorMessage error={getError('password')} />
                        </div>

                        <button
                            type="submit"
                            disabled={isSubmitting}
                            className="w-full bg-gradient-to-r from-stone-800 to-stone-900 active:from-stone-700 active:to-stone-800 text-white font-semibold py-3 md:py-3.5 rounded-lg md:rounded-xl text-sm md:text-base transition-all transform active:scale-[0.98] shadow-lg shadow-stone-900/25 flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed"
                        >
                            {isSubmitting ? <Loader2 className="w-5 h-5 animate-spin" /> : <>Criar Conta <ArrowRight className="w-4 h-4 md:w-5 md:h-5" /></>}
                        </button>
                    </form>

                    <div className="mt-6 md:mt-8 text-center">
                        <p className="text-stone-500 text-xs md:text-sm">
                            Já tem uma conta?{' '}
                            <Link to="/login" className="text-stone-800 active:text-stone-600 font-medium transition-colors hover:underline">
                                Faça login
                            </Link>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};
