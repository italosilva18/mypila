import React from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { BottomNavigation } from './BottomNavigation';

export const Layout: React.FC = () => {
    return (
        <div className="flex min-h-screen bg-paper">
            <Sidebar />
            <main className="flex-1 md:ml-64 px-2 py-3 md:p-8 overflow-y-auto h-screen">
                <Outlet />
            </main>
            <BottomNavigation />
        </div>
    );
};
