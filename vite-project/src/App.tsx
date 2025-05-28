import { useEffect, useState } from 'react';
import './App.css';
import { AuthProvider } from './AuthContext';
import Configurator from './components/Configurator/Configurator';
import Header from './components/Header/Header';
import { categories, CategoryType, Page } from './types';
import Account from './components/Account/Account';
import { ConfigProvider } from './ConfigContext';

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('configurator');


  return (
    <AuthProvider>
      <ConfigProvider>
        <Header setCurrentPage={setCurrentPage} currentPage={currentPage} />
        {currentPage === 'configurator' && <Configurator />}
        {currentPage === 'account' && <Account />}
      </ConfigProvider>
    </AuthProvider>
  );
}

export default App;
