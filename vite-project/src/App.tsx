import { useState } from 'react';
import './App.css';
import { AuthProvider } from './AuthContext';
import Configurator from './components/Configurator/Configurator';
import Header from './components/Header/Header';
import { Page } from './types';
import Account from './components/Account/Account';
import { ConfigProvider } from './ConfigContext';
import UsecasesPage from './components/UsecasesPage/UsecasesPage';
import './App.css';

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('configurator');

  return (
    <AuthProvider>
      <ConfigProvider>
        <Header setCurrentPage={setCurrentPage} currentPage={currentPage} />
        {currentPage === 'configurator' && <Configurator />}
        {currentPage === 'usecases' && <UsecasesPage />}
        {currentPage === 'account' && <Account />}
      </ConfigProvider>
    </AuthProvider>
  );
}

export default App;
