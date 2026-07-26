import React, { useState, useEffect } from 'react';
import {
	ThemeProvider,
	createTheme,
	CssBaseline,
	Box,
	AppBar,
	Toolbar,
	Typography,
	Button
} from '@mui/material';
import LogoutIcon from '@mui/icons-material/Logout';
import PodListExplorer from './pages/PodListExplorer';
import { Login } from './pages/Login';

const theme = createTheme({
	palette: {
		mode: 'dark',
		background: {
			default: '#0d1117',
			paper: '#161b22',
		},
	},
});

export default function App() {
	const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);

	useEffect(() => {
		const token = localStorage.getItem('authToken');
		if (token) {
			setIsAuthenticated(true);
		}
	}, []);

	const handleLoginSuccess = () => {
		setIsAuthenticated(true);
	};

	const handleLogout = () => {
		localStorage.removeItem('authToken');
		setIsAuthenticated(false);
	};

	return (
		<ThemeProvider theme={theme}>
			<CssBaseline />
			{!isAuthenticated ? (
				<Login onLoginSuccess={handleLoginSuccess} />
			) : (
				<Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
					<AppBar position="static" color="default" elevation={1}>
						<Toolbar variant="dense" sx={{ justifyContent: 'space-between' }}>
							<Typography variant="h6" component="div" sx={{ fontSize: '1rem', fontWeight: 600 }}>
								Go Through the Storm
							</Typography>
							<Button
								color="inherit"
								size="small"
								onClick={handleLogout}
								startIcon={<LogoutIcon />}
							>
								Logout
							</Button>
						</Toolbar>
					</AppBar>

					<Box sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column' }}>
						<PodListExplorer />
					</Box>
				</Box>
			)}
		</ThemeProvider>
	);
}