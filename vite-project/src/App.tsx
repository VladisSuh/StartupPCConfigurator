
import './App.css'
import { AuthProvider } from './AuthContext'
import Configurator from './components/Configurator/Configurator'
import Header from './components/Header/Header'

function App() {

  return (
    <>
      <AuthProvider>
        <Header />
        <Configurator />
      </AuthProvider>

    </>
  )
}

export default App
