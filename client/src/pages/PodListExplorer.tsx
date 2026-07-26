import React, { useState, useEffect } from 'react';
import {
	Box,
	Grid,
	Paper,
	List,
	ListItemButton,
	ListItemIcon,
	ListItemText,
	Typography,
	Chip,
	TextField,
	InputAdornment,
	Divider,
	IconButton,
	Stack,
	Tooltip,
	Breadcrumbs,
	Link,
	MenuItem,
	Select,
	FormControl,
	InputLabel,
	Tabs,
	Tab,
	SelectChangeEvent,
	CircularProgress,
	Alert,
	Button
} from '@mui/material';
import {
	Dns as PodIcon,
	Search as SearchIcon,
	Refresh as RefreshIcon,
	Terminal as LogsIcon,
	CheckCircle as RunningIcon,
	Error as ErrorIcon,
	Pending as PendingIcon,
	FolderZip as FileIcon,
	Download as DownloadIcon,
	ContentCopy as CopyIcon,
	ArrowBack as BackIcon,
	FolderOpen as OpenFolderIcon,
	Edit as EditIcon,
	Save as SaveIcon,
	Close as CancelIcon,
} from '@mui/icons-material';

import { useNamespaces } from '../hooks/useNamespaces';
import { usePods } from '../hooks/usePods';
import { fileApi } from '../api/fileApi';
import { ContainerFile } from '../models/k8sResource';
import { useFiles } from '../hooks/useFiles';
import { getFileIcon } from '../components/fileIcon';

export type PodStatus = 'Running' | 'Pending' | 'CrashLoopBackOff' | 'Completed' | 'Failed';

export interface ContainerInfo {
	name: string;
	image: string;
	restarts: number;
	ready: boolean;
	files?: ContainerFile[];
}

export interface PodItem {
	id: string;
	name: string;
	namespace: string;
	status: PodStatus;
	node: string;
	ip: string;
	age: string;
	cpuUsage: string;
	memoryUsage: string;
	containers: ContainerInfo[];
}

// Styling for standard light scrollbars
const scrollbarStyles = {
	overflowY: 'auto',
	'&::-webkit-scrollbar': {
		width: '8px',
		height: '8px',
	},
	'&::-webkit-scrollbar-track': {
		backgroundColor: 'rgba(0, 0, 0, 0.05)',
		borderRadius: '4px',
	},
	'&::-webkit-scrollbar-thumb': {
		backgroundColor: 'rgba(0, 0, 0, 0.25)',
		borderRadius: '4px',
		'&:hover': {
			backgroundColor: 'rgba(0, 0, 0, 0.4)',
		},
	},
};

// Styling for dark console / code viewer scrollbars
const darkScrollbarStyles = {
	overflowY: 'auto',
	'&::-webkit-scrollbar': {
		width: '8px',
		height: '8px',
	},
	'&::-webkit-scrollbar-track': {
		backgroundColor: '#090d11',
		borderRadius: '4px',
	},
	'&::-webkit-scrollbar-thumb': {
		backgroundColor: '#30363d',
		borderRadius: '4px',
		'&:hover': {
			backgroundColor: '#8b949e',
		},
	},
};

export const ContainerFileExplorerTab: React.FC<{
	podNamespace: string;
	podName: string;
	activeContainer?: ContainerInfo;
}> = ({ podNamespace, podName, activeContainer }) => {
	const [currentPath, setCurrentPath] = useState<string>('/');
	const [selectedFile, setSelectedFile] = useState<ContainerFile | null>(null);

	const {
		files,
		fileContent,
		loadingFiles,
		loadingContent,
		saving,
		error: fileError,
		fetchFiles,
		fetchFileContent,
		saveFile
	} = useFiles({
		namespace: podNamespace,
		pod: podName,
		container: activeContainer?.name || '',
		path: currentPath
	});

	useEffect(() => {
		if (activeContainer?.name) {
			fetchFiles();
		}
	}, [currentPath, activeContainer?.name, podName, podNamespace]);

	useEffect(() => {
		if (selectedFile && !selectedFile.isDirectory) {
			fetchFileContent(selectedFile.path);
			setIsEditing(false);
		}
	}, [selectedFile]);

	// File editor state
	const [isEditing, setIsEditing] = useState<boolean>(false);
	const [editedContent, setEditedContent] = useState<string>('');

	const handleCancelEdit = () => {
		const hasUnsavedChanges = editedContent !== (fileContent || '');

		if (hasUnsavedChanges) {
			const confirmCancel = window.confirm(
				'You have unsaved changes. Are you sure you want to discard them?'
			);
			if (!confirmCancel) return;
		}

		setEditedContent(fileContent || '');
		setIsEditing(false);
	};

	const handleEnterEditing = async () => {
		if (selectedFile) {
			const shouldDownloadBackup = window.confirm(
				`Would you like to download a backup of "${selectedFile.name}" before editing?`
			);

			if (shouldDownloadBackup) {
				await handleDownload(selectedFile);
			}
		}

		setIsEditing(true);
	};

	const handleSave = async () => {
		if (!selectedFile) return;
		const success = await saveFile(selectedFile.path, editedContent);
		if (success) {
			setIsEditing(false);
			fetchFileContent(selectedFile.path);
		}
	};

	const hasUnsavedChanges = isEditing && editedContent !== fileContent;

	useEffect(() => {
		setEditedContent(fileContent || '');
	}, [fileContent]);

	useEffect(() => {
		setCurrentPath('/');
		setSelectedFile(null);
		setIsEditing(false);
	}, [activeContainer?.name, podName, podNamespace]);

	const handleOpenFolder = (folderPath: string) => {
		const formattedPath = folderPath.endsWith('/') ? folderPath : `${folderPath}/`;
		setCurrentPath(formattedPath);
		setSelectedFile(null);
	};

	const handleDownload = async (file: ContainerFile) => {
		if (file.isDirectory || !activeContainer) return;

		await fileApi.downloadFile(
			podNamespace,
			podName,
			activeContainer.name,
			file.path,
			file.name
		);
	};

	return (
		<Grid container spacing={2} sx={{ flexGrow: 1, height: '100%', minHeight: 0, overflow: 'hidden' }}>
			{fileError && (
				<Grid size={{ xs: 12 }}>
					<Alert severity="error" sx={{ py: 0 }}>
						{fileError}
					</Alert>
				</Grid>
			)}

			{/* Directory Browser Panel - Fixed flex column to constrain height */}
			<Grid size={{ xs: 12, md: 5 }} sx={{ height: '100%', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
				<Paper variant="outlined" sx={{ p: 1, mb: 1, bgcolor: 'action.hover', flexShrink: 0 }}>
					<Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
						{currentPath !== '/' && (
							<IconButton
								size="small"
								onClick={() => {
									const segments = currentPath.split('/').filter(Boolean);
									segments.pop();
									const parentPath = segments.length > 0 ? `/${segments.join('/')}/` : '/';
									setCurrentPath(parentPath);
									setSelectedFile(null);
								}}
							>
								<BackIcon fontSize="small" />
							</IconButton>
						)}
						<OpenFolderIcon fontSize="small" color="action" />
						<Breadcrumbs maxItems={3} aria-label="breadcrumb" sx={{ fontSize: '0.8rem' }}>
							<Link
								underline="hover"
								color="inherit"
								sx={{ cursor: 'pointer' }}
								onClick={() => {
									setCurrentPath('/');
									setSelectedFile(null);
								}}
							>
								root
							</Link>
							{currentPath
								.split('/')
								.filter(Boolean)
								.map((folder, idx, arr) => {
									const subPath = '/' + arr.slice(0, idx + 1).join('/') + '/';
									return (
										<Link
											key={subPath}
											underline="hover"
											color={idx === arr.length - 1 ? 'text.primary' : 'inherit'}
											sx={{ cursor: 'pointer' }}
											onClick={() => {
												setCurrentPath(subPath);
												setSelectedFile(null);
											}}
										>
											{folder}
										</Link>
									);
								})}
						</Breadcrumbs>
						<IconButton size="small" onClick={fetchFiles} sx={{ ml: 'auto' }}>
							<RefreshIcon fontSize="small" />
						</IconButton>
					</Stack>
				</Paper>

				<Paper variant="outlined" sx={{ flexGrow: 1, minHeight: 0, ...scrollbarStyles }}>
					{loadingFiles ? (
						<Stack sx={{ alignItems: 'center', justifyContent: 'center', py: 4 }}>
							<CircularProgress size={24} />
							<Typography variant="caption" sx={{ mt: 1 }} color="text.secondary">
								Fetching folder content...
							</Typography>
						</Stack>
					) : (
						<List dense>
							{files && files.length > 0 ? (
								files.map((file, idx) => (
									<ListItemButton
										key={file.path || idx}
										selected={selectedFile?.path === file.path}
										onClick={() => {
											if (file.isDirectory) {
												handleOpenFolder(file.path);
											} else {
												setSelectedFile(file);
											}
										}}
									>
										<ListItemIcon sx={{ minWidth: 36 }}>{getFileIcon(file)}</ListItemIcon>
										<ListItemText
											primary={file.name}
											secondary={
												file.isDirectory
													? 'Directory'
													: `${file.size || '0 KB'} ${file.permissions ? `• ${file.permissions}` : ''}`
											}
											slotProps={{
												primary: {
													variant: 'body2',
													sx: {
														fontWeight: file.isDirectory ? 600 : 400,
													},
												},
											}}
										/>
									</ListItemButton>
								))
							) : (
								<Typography variant="caption" sx={{ p: 2, display: 'block', color: 'text.secondary' }}>
									No items inside <code>{currentPath}</code>
								</Typography>
							)}
						</List>
					)}
				</Paper>
			</Grid>

			{/* Code Viewer & Editor Panel */}
			<Grid size={{ xs: 12, md: 7 }} sx={{ height: '100%', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
				{selectedFile ? (
					<Paper variant="outlined" sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column', minHeight: 0, overflow: 'hidden' }}>
						<Box sx={{ flexShrink: 0 }}>
							<Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'flex-start', mb: 1 }}>
								<Box>
									<Stack direction="row" spacing={1} sx={{ alignItems: 'center', mb: 0.5 }}>
										<Typography variant="subtitle1" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
											{selectedFile.name}
										</Typography>
										{hasUnsavedChanges && (
											<Chip label="Unsaved Changes" size="small" color="warning" variant="outlined" />
										)}
									</Stack>
									<Typography variant="caption" color="text.secondary">
										Path: <code>{selectedFile.path}</code>
									</Typography>
								</Box>

								<Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
									{isEditing ? (
										<>
											<Button
												size="small"
												variant="outlined"
												color="inherit"
												startIcon={<CancelIcon />}
												onClick={handleCancelEdit}
												disabled={saving}
											>
												Cancel
											</Button>
											<Button
												size="small"
												variant="contained"
												color="primary"
												startIcon={saving ? <CircularProgress size={16} color="inherit" /> : <SaveIcon />}
												onClick={handleSave}
												disabled={saving}
											>
												Save
											</Button>
										</>
									) : (
										<>
											<Button
												size="small"
												variant="outlined"
												startIcon={<EditIcon />}
												onClick={() => handleEnterEditing()}
												disabled={loadingContent || !selectedFile.editable}
											>
												Edit
											</Button>

											<Tooltip title="Download File">
												<IconButton
													size="small"
													color="primary"
													onClick={() => handleDownload(selectedFile)}
													disabled={loadingContent}
												>
													<DownloadIcon fontSize="small" />
												</IconButton>
											</Tooltip>

											<Tooltip title="Copy File Path">
												<IconButton size="small" onClick={() => navigator.clipboard.writeText(selectedFile.path)}>
													<CopyIcon fontSize="small" />
												</IconButton>
											</Tooltip>
										</>
									)}
								</Stack>
							</Stack>

							<Stack direction="row" spacing={1} sx={{ mb: 1.5 }}>
								<Chip label={`Size: ${selectedFile.size || 'N/A'}`} size="small" variant="filled" />
								{selectedFile.permissions && (
									<Chip label={`Mode: ${selectedFile.permissions}`} size="small" variant="filled" />
								)}
								{selectedFile.updatedAt && (
									<Chip label={`Modified: ${selectedFile.updatedAt}`} size="small" variant="filled" />
								)}
								<Chip
									label={selectedFile.editable ? "Editable" : "Read-Only"}
									color={selectedFile.editable ? "success" : "error"}
									size="small"
									variant="filled"
								/>
							</Stack>

							<Divider sx={{ mb: 1.5 }} />
						</Box>

						{/* Scrollable code window with fixed parent constraints */}
						<Box
							component="pre"
							sx={{
								p: 1.5,
								m: 0,
								flexGrow: 1,
								minHeight: 0,
								borderRadius: 1,
								bgcolor: '#0d1117',
								color: '#c9d1d9',
								fontFamily: 'monospace',
								fontSize: '0.8rem',
								border: '1px solid',
								borderColor: 'divider',
								...darkScrollbarStyles
							}}
						>
							{loadingContent ? (
								<Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
									<CircularProgress size={16} color="inherit" />
									<Typography variant="caption">Loading file content...</Typography>
								</Stack>
							) : isEditing ? (
								<TextField
									multiline
									fullWidth
									value={editedContent}
									onChange={(e) => setEditedContent(e.target.value)}
									variant="outlined"
									slotProps={{
										input: {
											sx: {
												fontFamily: 'monospace',
												fontSize: '0.85rem',
												bgcolor: '#0d1117',
												color: '#c9d1d9',
												height: '100%',
												alignItems: 'flex-start',
												'& textarea': {
													height: '100% !important',
													overflowY: 'auto !important',
													...darkScrollbarStyles
												}
											}
										}
									}}
									sx={{ height: '100%', flexGrow: 1, display: 'flex', flexDirection: 'column' }}
								/>
							) : (
								<code>{fileContent || selectedFile.content || 'File is empty.'}</code>
							)}
						</Box>
					</Paper>
				) : (
					<Paper
						variant="outlined"
						sx={{
							p: 2,
							height: '100%',
							display: 'flex',
							alignItems: 'center',
							justifyContent: 'center',
							color: 'text.secondary'
						}}
					>
						<Typography variant="caption">Select a file from the list to preview details & download</Typography>
					</Paper>
				)}
			</Grid>
		</Grid>
	);
};

export const PodListExplorer: React.FC = () => {
	const { namespaces, selectedNs, setSelectedNs, loading: loadingNs, error: nsError } = useNamespaces();

	const [searchQuery, setSearchQuery] = useState<string>('');
	const [namespaceFilter, setNamespaceFilter] = useState<string>('all');

	const {
		pods,
		selectedPod,
		setSelectedPod,
		logs,
		loading: loadingPods,
		logsLoading,
		error: podsError,
		fetchPods,
		fetchLogs
	} = usePods(namespaceFilter);

	// Detail Panel State
	const [activeTab, setActiveTab] = useState<number>(0);
	const [selectedContainerIndex, setSelectedContainerIndex] = useState<number>(0);

	useEffect(() => {
		if (selectedNs) {
			setNamespaceFilter(selectedNs);
		}
	}, [selectedNs]);

	const activeContainer = selectedPod?.containers[selectedContainerIndex];

	useEffect(() => {
		if (selectedPod && activeContainer && activeTab === 0) {
			fetchLogs(selectedPod.namespace, selectedPod.name, activeContainer.name);
		}
	}, [selectedPod, selectedContainerIndex, activeTab]);

	const handleNamespaceChange = (e: SelectChangeEvent) => {
		const val = e.target.value;
		setNamespaceFilter(val);
		if (val !== 'all') {
			setSelectedNs(val);
		}
	};

	const filteredPods = (pods || []).filter((pod) =>
		pod.name.toLowerCase().includes(searchQuery.toLowerCase())
	);

	const getStatusChip = (status: PodStatus): React.ReactElement => {
		switch (status) {
			case 'Running':
				return <Chip icon={<RunningIcon />} label="Running" color="success" size="small" variant="outlined" />;
			case 'Pending':
				return <Chip icon={<PendingIcon />} label="Pending" color="warning" size="small" variant="outlined" />;
			case 'CrashLoopBackOff':
			case 'Failed':
				return <Chip icon={<ErrorIcon />} label={status} color="error" size="small" variant="outlined" />;
			default:
				return <Chip label={status} size="small" />;
		}
	};

	const displayError = nsError || podsError;

	return (
		<Box sx={{ p: 3, bgcolor: 'background.default', height: '100vh', overflow: 'hidden', boxSizing: 'border-box' }}>
			<Grid container spacing={2} sx={{ height: '100%' }}>
				{/* Left Sidebar Panel - Pod List */}
				<Grid size={{ xs: 12, md: 5, lg: 4 }} sx={{ height: '100%', minHeight: 0 }}>
					<Paper
						elevation={0}
						variant="outlined"
						sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column', boxSizing: 'border-box' }}
					>
						<Box sx={{ flexShrink: 0 }}>
							<Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
								<IconButton size="small" aria-label="refresh list" onClick={fetchPods}>
									<RefreshIcon fontSize="small" />
								</IconButton>
							</Stack>

							{displayError && (
								<Alert severity="error" sx={{ mb: 2, py: 0 }}>
									{displayError}
								</Alert>
							)}

							<Stack direction="row" spacing={1} sx={{ mb: 2 }}>
								<TextField
									size="small"
									placeholder="Search pod name..."
									value={searchQuery}
									onChange={(e) => setSearchQuery(e.target.value)}
									fullWidth
									slotProps={{
										input: {
											startAdornment: (
												<InputAdornment position="start">
													<SearchIcon fontSize="small" />
												</InputAdornment>
											)
										}
									}}
								/>
								<FormControl size="small" sx={{ minWidth: 140 }}>
									<InputLabel id="namespace-select-label">Namespace</InputLabel>
									<Select
										labelId="namespace-select-label"
										value={namespaceFilter}
										label="Namespace"
										onChange={handleNamespaceChange}
										disabled={loadingNs}
										startAdornment={
											loadingNs ? (
												<InputAdornment position="start">
													<CircularProgress size={16} />
												</InputAdornment>
											) : null
										}
									>
										<MenuItem value="all">All Namespaces</MenuItem>
										{namespaces.map((ns) => (
											<MenuItem key={ns.name} value={ns.name}>
												{ns.name}
											</MenuItem>
										))}
									</Select>
								</FormControl>
							</Stack>

							<Divider sx={{ mb: 1 }} />
						</Box>

						{/* Pod list container set to flex-grow with strict min-height */}
						<List sx={{ flexGrow: 1, minHeight: 0, pr: 0.5, ...scrollbarStyles }}>
							{loadingPods ? (
								<Stack sx={{ alignItems: 'center', justifyContent: 'center', py: 4 }}>
									<CircularProgress size={28} />
									<Typography variant="caption" sx={{ mt: 1 }} color="text.secondary">
										Loading pods...
									</Typography>
								</Stack>
							) : filteredPods && filteredPods.length > 0 ? (
								filteredPods.map((pod) => (
									<ListItemButton
										key={pod.id}
										selected={selectedPod?.id === pod.id}
										onClick={() => {
											setSelectedPod(pod);
											setSelectedContainerIndex(0);
										}}
										sx={{ borderRadius: 1.5, mb: 1, flexDirection: 'column', alignItems: 'flex-start' }}
									>
										<Stack direction="row" sx={{ justifyContent: 'space-between', width: '100%', mb: 0.5 }}>
											<Typography variant="body2" sx={{ fontWeight: 600, noWrap: true, maxWidth: '65%' }}>
												{pod.name}
											</Typography>
											{getStatusChip(pod.status)}
										</Stack>
										<Typography variant="caption" color="text.secondary">
											{pod.namespace} • Age: {pod.age}
										</Typography>
									</ListItemButton>
								))
							) : (
								<Typography variant="caption" sx={{ p: 2, display: 'block', color: 'text.secondary' }}>
									No pods found.
								</Typography>
							)}
						</List>
					</Paper>
				</Grid>

				{/* Right Main Details Panel */}
				<Grid size={{ xs: 12, md: 7, lg: 8 }} sx={{ height: '100%', minHeight: 0 }}>
					<Paper
						elevation={0}
						variant="outlined"
						sx={{ p: 3, height: '100%', display: 'flex', flexDirection: 'column', boxSizing: 'border-box' }}
					>
						{selectedPod ? (
							<>
								<Box sx={{ flexShrink: 0 }}>
									<Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
										<Box>
											<Stack direction="row" spacing={1} sx={{ alignItems: 'center', mb: 0.5 }}>
												<PodIcon color="primary" />
												<Typography variant="h6" sx={{ fontWeight: 600 }}>
													{selectedPod.name}
												</Typography>
											</Stack>
											<Typography variant="caption" color="text.secondary">
												Namespace: <strong>{selectedPod.namespace}</strong> | IP: <strong>{selectedPod.ip}</strong>
											</Typography>
										</Box>
										{getStatusChip(selectedPod.status)}
									</Stack>

									<Stack direction="row" spacing={2} sx={{ alignItems: 'center', mb: 2 }}>
										<FormControl size="small" sx={{ minWidth: 200 }}>
											<InputLabel id="container-select-label">Select Container</InputLabel>
											<Select
												labelId="container-select-label"
												value={selectedContainerIndex.toString()}
												label="Select Container"
												onChange={(e: SelectChangeEvent) => setSelectedContainerIndex(Number(e.target.value))}
											>
												{selectedPod.containers?.map((c, idx) => (
													<MenuItem key={idx} value={idx}>
														{c.name}
													</MenuItem>
												))}
											</Select>
										</FormControl>

										<Tabs value={activeTab} onChange={(_, val) => setActiveTab(val)}>
											<Tab icon={<LogsIcon fontSize="small" />} iconPosition="start" label="Console Logs" />
											<Tab icon={<FileIcon fontSize="small" />} iconPosition="start" label="File Explorer" />
										</Tabs>
									</Stack>

									<Divider sx={{ mb: 2 }} />
								</Box>

								{/* Console Logs Tab Content */}
								{activeTab === 0 && (
									<Box
										component="pre"
										sx={{
											p: 2,
											m: 0,
											flexGrow: 1,
											minHeight: 0,
											borderRadius: 1.5,
											bgcolor: '#0d1117',
											color: '#c9d1d9',
											fontFamily: 'monospace',
											fontSize: '0.8rem',
											border: '1px solid',
											borderColor: 'divider',
											...darkScrollbarStyles
										}}
									>
										{logsLoading ? (
											<Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
												<CircularProgress size={16} color="inherit" />
												<Typography variant="caption">Fetching logs...</Typography>
											</Stack>
										) : (
											<code>{logs || 'No logs available for this container.'}</code>
										)}
									</Box>
								)}

								{/* File Explorer Tab Content */}
								{activeTab === 1 && (
									<Box sx={{ flexGrow: 1, minHeight: 0, height: '100%' }}>
										<ContainerFileExplorerTab
											podNamespace={selectedPod.namespace}
											podName={selectedPod.name}
											activeContainer={activeContainer}
										/>
									</Box>
								)}
							</>
						) : (
							<Stack sx={{ alignItems: 'center', justifyContent: 'center', height: '100%', color: 'text.secondary' }}>
								<PodIcon sx={{ fontSize: 48, mb: 1, opacity: 0.5 }} />
								<Typography>Select a Pod to view details</Typography>
							</Stack>
						)}
					</Paper>
				</Grid>
			</Grid>
		</Box>
	);
};

export default PodListExplorer;