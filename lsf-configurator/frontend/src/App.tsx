import { Routes, Route } from 'react-router'
import Home from './pages/Home'
import FunctionAppDetails from './pages/FunctionAppDetails'
import './App.css'

function App() {

  return (
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/function-apps/:id" element={<FunctionAppDetails />} />
      </Routes>
  )
}

export default App
