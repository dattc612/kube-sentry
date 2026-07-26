import React, { useState } from 'react';
import { Box, Button, TextField, Paper, Typography, Alert } from '@mui/material';
import axios from 'axios';
import { authApi } from '../api/authApi';

export const Login: React.FC<{ onLoginSuccess: () => void }> = ({ onLoginSuccess }) => {
	const [username, setUsername] = useState('');
	const [password, setPassword] = useState('');
	const [error, setError] = useState('');

	const handleLogin = async (e: React.FormEvent) => {
		e.preventDefault();
		try {
			const res = await authApi.login(username, password);
			localStorage.setItem('authToken', res.token);
			onLoginSuccess();
		} catch (err: any) {
			setError(err.response?.data?.error || 'Failed to authenticate');
		}
	};

	return (
		<Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', bgcolor: '#0d1117' }}>
			<Paper variant="outlined" sx={{ p: 4, width: 320, display: 'flex', flexDirection: 'column', gap: 2 }}>
				<Typography variant="h6" sx={{ fontWeight: 700 }}>Sign In</Typography>
				{error && <Alert severity="error">{error}</Alert>}
				<form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
					<TextField label="Username" size="small" value={username} onChange={(e) => setUsername(e.target.value)} fullWidth />
					<TextField label="Password" type="password" size="small" value={password} onChange={(e) => setPassword(e.target.value)} fullWidth />
					<Button type="submit" variant="contained" fullWidth>Login</Button>
				</form>
			</Paper>
		</Box>
	);
};