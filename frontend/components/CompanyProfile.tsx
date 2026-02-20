import React, { useState, useEffect, useCallback, useRef } from 'react';
import { X, Save, Search, Loader2, Building2, Image, Upload } from 'lucide-react';
import { Company, UpdateCompanyRequest } from '../types';
import { api } from '../services/api';
import { useFormValidation } from '../hooks/useFormValidation';
import { useEscapeKey } from '../hooks/useEscapeKey';
import { validateRequired, validateMaxLength, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';
import { useToast } from '../contexts/ToastContext';

interface Props {
    isOpen: boolean;
    onClose: () => void;
    onSave: (company: Company) => void;
    company: Company;
}

export const CompanyProfile: React.FC<Props> = ({ isOpen, onClose, onSave, company }) => {
    const { addToast } = useToast();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isSearchingCNPJ, setIsSearchingCNPJ] = useState(false);
    const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

    // Form fields
    const [name, setName] = useState('');
    const [cnpj, setCnpj] = useState('');
    const [legalName, setLegalName] = useState('');
    const [tradeName, setTradeName] = useState('');
    const [email, setEmail] = useState('');
    const [phone, setPhone] = useState('');
    const [address, setAddress] = useState('');
    const [city, setCity] = useState('');
    const [state, setState] = useState('');
    const [zipCode, setZipCode] = useState('');
    const [logoUrl, setLogoUrl] = useState('');
    const [isUploadingLogo, setIsUploadingLogo] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleClose = useCallback(() => {
        clearAllErrors();
        onClose();
    }, [clearAllErrors, onClose]);

    useEscapeKey(handleClose, isOpen);

    useEffect(() => {
        if (company && isOpen) {
            setName(company.name || '');
            setCnpj(company.cnpj || '');
            setLegalName(company.legalName || '');
            setTradeName(company.tradeName || '');
            setEmail(company.email || '');
            setPhone(company.phone || '');
            setAddress(company.address || '');
            setCity(company.city || '');
            setState(company.state || '');
            setZipCode(company.zipCode || '');
            setLogoUrl(company.logoUrl || '');
            clearAllErrors();
        }
    }, [company, isOpen, clearAllErrors]);

    // Formata CPF (11 digitos) ou CNPJ (14 digitos)
    const formatCPFCNPJ = (value: string): string => {
        const digits = value.replace(/\D/g, '').slice(0, 14);

        // CPF: 000.000.000-00 (11 digitos)
        if (digits.length <= 11) {
            if (digits.length <= 3) return digits;
            if (digits.length <= 6) return `${digits.slice(0, 3)}.${digits.slice(3)}`;
            if (digits.length <= 9) return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6)}`;
            return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6, 9)}-${digits.slice(9)}`;
        }

        // CNPJ: 00.000.000/0000-00 (14 digitos)
        if (digits.length <= 12) return `${digits.slice(0, 2)}.${digits.slice(2, 5)}.${digits.slice(5, 8)}/${digits.slice(8)}`;
        return `${digits.slice(0, 2)}.${digits.slice(2, 5)}.${digits.slice(5, 8)}/${digits.slice(8, 12)}-${digits.slice(12)}`;
    };

    const formatPhone = (value: string): string => {
        const digits = value.replace(/\D/g, '').slice(0, 11);
        if (digits.length <= 2) return digits;
        if (digits.length <= 6) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
        if (digits.length <= 10) return `(${digits.slice(0, 2)}) ${digits.slice(2, 6)}-${digits.slice(6)}`;
        return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7)}`;
    };

    const formatZipCode = (value: string): string => {
        const digits = value.replace(/\D/g, '').slice(0, 8);
        if (digits.length <= 5) return digits;
        return `${digits.slice(0, 5)}-${digits.slice(5)}`;
    };

    const handleCNPJChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setCnpj(formatCPFCNPJ(e.target.value));
    };

    const handlePhoneChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setPhone(formatPhone(e.target.value));
    };

    const handleZipCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setZipCode(formatZipCode(e.target.value));
    };

    const handleLogoUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        // Validate file type
        if (!file.type.startsWith('image/')) {
            addToast('error', 'Por favor, selecione uma imagem');
            return;
        }

        // Validate file size (max 500KB for base64)
        if (file.size > 500 * 1024) {
            addToast('error', 'Imagem muito grande. Maximo 500KB');
            return;
        }

        setIsUploadingLogo(true);
        try {
            // Convert to base64
            const reader = new FileReader();
            reader.onloadend = () => {
                const base64 = reader.result as string;
                setLogoUrl(base64);
                setIsUploadingLogo(false);
                addToast('success', 'Logo carregada com sucesso!');
            };
            reader.onerror = () => {
                addToast('error', 'Erro ao carregar imagem');
                setIsUploadingLogo(false);
            };
            reader.readAsDataURL(file);
        } catch {
            addToast('error', 'Erro ao processar imagem');
            setIsUploadingLogo(false);
        }
    };

    const handleRemoveLogo = () => {
        setLogoUrl('');
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
    };

    const handleSearchCNPJ = async () => {
        const cleanCNPJ = cnpj.replace(/\D/g, '');
        if (cleanCNPJ.length !== 14) {
            addToast('error', 'CNPJ deve ter 14 digitos');
            return;
        }

        try {
            setIsSearchingCNPJ(true);
            const data = await api.lookupCNPJ(cleanCNPJ);

            // Auto-fill fields with CNPJ data
            setLegalName(data.razaoSocial || '');
            setTradeName(data.nomeFantasia || '');

            // Build address from parts
            let fullAddress = data.logradouro || '';
            if (data.numero) fullAddress += `, ${data.numero}`;
            if (data.complemento) fullAddress += ` - ${data.complemento}`;
            if (data.bairro) fullAddress += ` - ${data.bairro}`;
            setAddress(fullAddress);

            setCity(data.municipio || '');
            setState(data.uf || '');
            setZipCode(formatZipCode(data.cep || ''));

            if (data.telefone) {
                setPhone(formatPhone(data.telefone));
            }

            // Use trade name as company name if name is empty
            if (!name && data.nomeFantasia) {
                setName(data.nomeFantasia);
            }

            addToast('success', 'Dados do CNPJ carregados com sucesso!');
        } catch (err) {
            console.error('CNPJ lookup failed:', err);
            addToast('error', 'Erro ao buscar dados do CNPJ. Verifique se o CNPJ esta correto.');
        } finally {
            setIsSearchingCNPJ(false);
        }
    };

    const validateForm = useCallback((): boolean => {
        return validateFields({
            name: () => combineValidations(
                validateRequired(name, 'Nome'),
                validateMaxLength(name, 100, 'Nome')
            )
        });
    }, [validateFields, name]);

    const handleSubmit = useCallback(async (e: React.FormEvent) => {
        e.preventDefault();
        if (!validateForm()) return;

        try {
            setIsSubmitting(true);

            const data: UpdateCompanyRequest = {
                name: name.trim(),
                cnpj: cnpj.replace(/\D/g, '') || undefined,
                legalName: legalName.trim() || undefined,
                tradeName: tradeName.trim() || undefined,
                email: email.trim() || undefined,
                phone: phone.replace(/\D/g, '') || undefined,
                address: address.trim() || undefined,
                city: city.trim() || undefined,
                state: state || undefined,
                zipCode: zipCode.replace(/\D/g, '') || undefined,
                logoUrl: logoUrl.trim() || undefined,
            };

            const updated = await api.updateCompany(company.id, data);
            addToast('success', 'Perfil atualizado com sucesso!');
            onSave(updated);
            onClose();
        } catch (err) {
            console.error('Failed to update company:', err);
            addToast('error', 'Erro ao atualizar perfil');
        } finally {
            setIsSubmitting(false);
        }
    }, [validateForm, name, cnpj, legalName, tradeName, email, phone, address, city, state, zipCode, logoUrl, company.id, addToast, onSave, onClose]);

    if (!isOpen) return null;

    const states = ['AC', 'AL', 'AP', 'AM', 'BA', 'CE', 'DF', 'ES', 'GO', 'MA', 'MT', 'MS', 'MG', 'PA', 'PB', 'PR', 'PE', 'PI', 'RJ', 'RN', 'RS', 'RO', 'RR', 'SC', 'SP', 'SE', 'TO'];

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-0 md:p-4 bg-stone-900/50 backdrop-blur-sm">
            <div
                role="dialog"
                aria-modal="true"
                aria-labelledby="company-profile-title"
                className="bg-white border-0 md:border border-stone-200 rounded-none md:rounded-2xl w-full h-full md:h-auto md:max-w-2xl md:max-h-[90vh] shadow-2xl overflow-hidden flex flex-col"
            >
                {/* Header */}
                <div className="flex justify-between items-center p-4 md:p-6 border-b border-stone-200 bg-stone-50 sticky top-0 z-10">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-stone-200 rounded-xl">
                            <Building2 className="w-5 h-5 text-stone-600" />
                        </div>
                        <h3 id="company-profile-title" className="text-lg md:text-xl font-bold text-stone-900">
                            Perfil
                        </h3>
                    </div>
                    <button
                        onClick={handleClose}
                        className="text-stone-400 hover:text-stone-700 transition-colors p-2 -mr-2 rounded-lg"
                        aria-label="Fechar modal"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-4 md:p-6 space-y-6">
                    {/* Logo Section */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Logo</h4>
                        <div className="flex items-start gap-4">
                            <div className="w-20 h-20 bg-stone-100 rounded-xl border-2 border-dashed border-stone-300 flex items-center justify-center overflow-hidden relative">
                                {logoUrl ? (
                                    <img
                                        src={logoUrl}
                                        alt="Logo"
                                        className="w-full h-full object-contain"
                                        onError={(e) => {
                                            (e.target as HTMLImageElement).style.display = 'none';
                                        }}
                                    />
                                ) : (
                                    <Image className="w-8 h-8 text-stone-400" />
                                )}
                                {isUploadingLogo && (
                                    <div className="absolute inset-0 bg-white/80 flex items-center justify-center">
                                        <Loader2 className="w-6 h-6 animate-spin text-stone-600" />
                                    </div>
                                )}
                            </div>
                            <div className="flex-1">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Logo da Empresa
                                </label>
                                <input
                                    ref={fileInputRef}
                                    type="file"
                                    accept="image/*"
                                    onChange={handleLogoUpload}
                                    className="hidden"
                                    id="logo-upload"
                                />
                                <div className="flex gap-2">
                                    <button
                                        type="button"
                                        onClick={() => fileInputRef.current?.click()}
                                        disabled={isUploadingLogo}
                                        className="flex-1 px-4 py-2.5 bg-stone-800 text-white rounded-xl hover:bg-stone-700 disabled:bg-stone-300 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                                    >
                                        {isUploadingLogo ? (
                                            <Loader2 className="w-4 h-4 animate-spin" />
                                        ) : (
                                            <Upload className="w-4 h-4" />
                                        )}
                                        <span>Enviar Logo</span>
                                    </button>
                                    {logoUrl && (
                                        <button
                                            type="button"
                                            onClick={handleRemoveLogo}
                                            className="px-4 py-2.5 bg-red-100 text-red-600 rounded-xl hover:bg-red-200 transition-colors"
                                        >
                                            <X className="w-4 h-4" />
                                        </button>
                                    )}
                                </div>
                                <p className="text-xs text-stone-500 mt-1">PNG, JPG ou SVG. Maximo 500KB. Aparece nos orcamentos PDF.</p>
                            </div>
                        </div>
                    </section>

                    {/* Basic Info */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Informacoes Basicas</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="md:col-span-2">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Nome <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    placeholder="Nome que aparece no sistema"
                                    className={`w-full px-3 py-2.5 bg-stone-50 border ${hasError('name') ? 'border-red-500' : 'border-stone-200'} rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400`}
                                />
                                <ErrorMessage error={getError('name')} />
                            </div>

                            <div className="md:col-span-2">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    CPF/CNPJ
                                </label>
                                <div className="flex gap-2">
                                    <input
                                        type="text"
                                        value={cnpj}
                                        onChange={handleCNPJChange}
                                        placeholder="000.000.000-00 ou 00.000.000/0000-00"
                                        className="flex-1 px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                    />
                                    <button
                                        type="button"
                                        onClick={handleSearchCNPJ}
                                        disabled={isSearchingCNPJ || cnpj.replace(/\D/g, '').length !== 14}
                                        className="px-4 py-2.5 bg-stone-800 text-white rounded-xl hover:bg-stone-700 disabled:bg-stone-300 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
                                    >
                                        {isSearchingCNPJ ? (
                                            <Loader2 className="w-4 h-4 animate-spin" />
                                        ) : (
                                            <Search className="w-4 h-4" />
                                        )}
                                        <span className="hidden md:inline">Buscar</span>
                                    </button>
                                </div>
                                <p className="text-xs text-stone-500 mt-1">Digite o CNPJ (14 digitos) e clique em Buscar para preencher automaticamente</p>
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Razao Social
                                </label>
                                <input
                                    type="text"
                                    value={legalName}
                                    onChange={(e) => setLegalName(e.target.value)}
                                    placeholder="Razao social completa"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Nome Fantasia
                                </label>
                                <input
                                    type="text"
                                    value={tradeName}
                                    onChange={(e) => setTradeName(e.target.value)}
                                    placeholder="Nome fantasia"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>
                        </div>
                    </section>

                    {/* Contact Info */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Contato</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Email
                                </label>
                                <input
                                    type="email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    placeholder="email@exemplo.com"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Telefone
                                </label>
                                <input
                                    type="text"
                                    value={phone}
                                    onChange={handlePhoneChange}
                                    placeholder="(00) 00000-0000"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>
                        </div>
                    </section>

                    {/* Address */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Endereco</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    CEP
                                </label>
                                <input
                                    type="text"
                                    value={zipCode}
                                    onChange={handleZipCodeChange}
                                    placeholder="00000-000"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div className="md:col-span-2">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Endereco
                                </label>
                                <input
                                    type="text"
                                    value={address}
                                    onChange={(e) => setAddress(e.target.value)}
                                    placeholder="Rua, numero, bairro"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Cidade
                                </label>
                                <input
                                    type="text"
                                    value={city}
                                    onChange={(e) => setCity(e.target.value)}
                                    placeholder="Cidade"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Estado
                                </label>
                                <select
                                    value={state}
                                    onChange={(e) => setState(e.target.value)}
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                >
                                    <option value="">Selecione</option>
                                    {states.map(s => (
                                        <option key={s} value={s}>{s}</option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </section>
                </form>

                {/* Footer */}
                <div className="p-4 md:p-6 border-t border-stone-200 bg-stone-50">
                    <div className="flex gap-3">
                        <button
                            type="button"
                            onClick={handleClose}
                            className="flex-1 py-3 text-stone-600 bg-stone-200 hover:bg-stone-300 rounded-xl font-medium transition-colors"
                        >
                            Cancelar
                        </button>
                        <button
                            onClick={handleSubmit}
                            disabled={isSubmitting || hasErrors()}
                            className="flex-1 py-3 text-white bg-stone-800 hover:bg-stone-700 rounded-xl font-medium transition-colors flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {isSubmitting ? (
                                <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                                <Save className="w-4 h-4" />
                            )}
                            Salvar Alteracoes
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};
