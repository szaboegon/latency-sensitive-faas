import React from "react"
import { BrowserRouter, Routes, Route } from "react-router"
import { Box } from "@mui/material"
import Home from "./pages/Home"
import FunctionAppDetails from "./pages/FunctionAppDetails"
import Sidebar from "./components/Sidebar"
import "./App.css"

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Box sx={{ display: "flex" }}>
        <Sidebar />
        <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/function-apps/:id" element={<FunctionAppDetails />} />
          </Routes>
        </Box>
      </Box>
    </BrowserRouter>
  )
}

export default App
