import React from 'react';
import { Outlet } from 'react-router';

export const Layout: React.FC = () => {
  return (
    <div className="flex flex-col min-h-screen">
      <header className="p-4 bg-gray-800 text-white text-center">Header</header>
      <main className="flex-grow p-4">
        <Outlet />
      </main>
      <footer className="p-4 bg-gray-800 text-white text-center">Footer</footer>
    </div>
  );
};
