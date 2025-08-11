import React, { useState } from "react"
import { BrowserRouter, Routes, Route } from "react-router"
import { Box, Fab } from "@mui/material"
import Home from "./pages/Home"
import FunctionAppDetails from "./pages/FunctionAppDetails"
import Sidebar from "./components/Sidebar"
import AddIcon from "@mui/icons-material/Add"
import AddFunctionAppModal from "./components/AddFunctionAppModal"

const App: React.FC = () => {
  const [isModalOpen, setModalOpen] = useState(false)

  const handleOpenModal = () => setModalOpen(true)
  const handleCloseModal = () => setModalOpen(false)

  return (
    <BrowserRouter>
      <Box sx={{ display: "flex" }}>
        <Sidebar />
        <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/function-apps/:id" element={<FunctionAppDetails />} />
          </Routes>
          <Fab
            color="primary"
            sx={{ position: "fixed", bottom: 16, right: 16 }}
            onClick={handleOpenModal}
          >
            <AddIcon />
          </Fab>
          <AddFunctionAppModal open={isModalOpen} onClose={handleCloseModal} />
        </Box>
      </Box>
    </BrowserRouter>
  )
}

export default App
