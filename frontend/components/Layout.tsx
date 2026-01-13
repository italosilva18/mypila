import React, { useState, useEffect, useCallback } from 'react';
import { Outlet, useParams } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { BottomNavigation } from './BottomNavigation';
import { CompanyProfile } from './CompanyProfile';
import { Company } from '../types';
import { api } from '../services/api';

export const Layout: React.FC = () => {
    const { companyId } = useParams<{ companyId: string }>();
    const [company, setCompany] = useState<Company | null>(null);
    const [isProfileOpen, setIsProfileOpen] = useState(false);

    useEffect(() => {
        const loadCompany = async () => {
            if (!companyId) return;
            try {
                const companies = await api.getCompanies();
                const found = companies.find(c => c.id === companyId);
                if (found) setCompany(found);
            } catch (err) {
                console.error('Failed to load company', err);
            }
        };
        loadCompany();
    }, [companyId]);

    const handleSettingsClick = useCallback(() => {
        setIsProfileOpen(true);
    }, []);

    const handleProfileSave = useCallback((updatedCompany: Company) => {
        setCompany(updatedCompany);
    }, []);

    return (
        <div className="flex min-h-screen bg-paper">
            <Sidebar onSettingsClick={company ? handleSettingsClick : undefined} />
            <main className="flex-1 md:ml-64 px-2 py-3 md:p-8 overflow-y-auto h-screen">
                <Outlet />
            </main>
            <BottomNavigation onSettingsClick={company ? handleSettingsClick : undefined} />

            {company && (
                <CompanyProfile
                    isOpen={isProfileOpen}
                    onClose={() => setIsProfileOpen(false)}
                    onSave={handleProfileSave}
                    company={company}
                />
            )}
        </div>
    );
};
