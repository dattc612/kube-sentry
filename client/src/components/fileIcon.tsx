import React from 'react';


import { ContainerFile } from '../models/k8sResource';
import {
	Folder as FolderIcon,
	InsertDriveFile as GenericDocIcon,
	Code as CodeIcon,
	Description as TextIcon,
	DataObject as JsonIcon,
	Terminal as ScriptIcon,
	FolderZip as ArchiveIcon,
	Image as ImageIcon,
	PictureAsPdf as PdfIcon,
	Storage as DbIcon,
	VideoLibrary as VideoIcon,
	AudioFile as AudioIcon,
	Settings as ConfigIcon,
	Build as BuildIcon,
	Security as KeyIcon,
	TableChart as SpreadsheetIcon,
	Lock as LockIcon,
} from '@mui/icons-material';

export function getFileIcon(file: ContainerFile) {
	if (file.isDirectory) return <FolderIcon sx={{ color: '#ffa726' }} />;

	const ext = file.name.split('.').pop()?.toLowerCase();
	switch (ext) {
		case 'js':
		case 'jsx':
		case 'ts':
		case 'tsx':
		case 'py':
		case 'go':
		case 'java':
		case 'c':
		case 'cpp':
		case 'cs':
		case 'rs':
		case 'rb':
		case 'php':
		case 'kt':
		case 'swift':
		case 'html':
		case 'htm':
		case 'css':
		case 'scss':
		case 'sass':
		case 'less':
		case 'vue':
		case 'svelte':
		case 'sql':
		case 'graphql':
		case 'gql':
		case 'proto':
			return <CodeIcon color="primary" />;

		case 'json':
		case 'json5':
		case 'yaml':
		case 'yml':
		case 'xml':
		case 'toml':
		case 'plist':
			return <JsonIcon color="warning" />;

		case 'sh':
		case 'bash':
		case 'zsh':
		case 'fish':
		case 'ps1':
		case 'bat':
		case 'cmd':
		case 'exe':
		case 'bin':
			return <ScriptIcon color="secondary" />;

		case 'txt':
		case 'log':
		case 'md':
		case 'markdown':
		case 'rst':
		case 'doc':
		case 'docx':
		case 'rtf':
			return <TextIcon color="action" />;

		case 'pdf':
			return <PdfIcon sx={{ color: '#e53935' }} />;

		case 'csv':
		case 'tsv':
		case 'xls':
		case 'xlsx':
		case 'ods':
			return <SpreadsheetIcon sx={{ color: '#2e7d32' }} />;

		case 'env':
		case 'config':
		case 'conf':
		case 'ini':
		case 'properties':
		case 'editorconfig':
		case 'gitignore':
		case 'dockerignore':
			return <ConfigIcon color="action" />;

		case 'dockerfile':
		case 'containerfile':
		case 'makefile':
		case 'gradle':
		case 'lock':
			return <BuildIcon color="action" />;

		case 'pem':
		case 'crt':
		case 'cer':
		case 'key':
		case 'pub':
		case 'gpg':
		case 'asc':
			return <KeyIcon sx={{ color: '#f57c00' }} />;

		case 'zip':
		case 'tar':
		case 'gz':
		case 'tgz':
		case 'bz2':
		case '7z':
		case 'rar':
		case 'xz':
		case 'iso':
		case 'deb':
		case 'rpm':
		case 'apk':
		case 'jar':
			return <ArchiveIcon sx={{ color: '#8d6e63' }} />;

		case 'png':
		case 'jpg':
		case 'jpeg':
		case 'gif':
		case 'svg':
		case 'webp':
		case 'ico':
		case 'bmp':
		case 'tiff':
		case 'psd':
		case 'ai':
			return <ImageIcon color="info" />;

		case 'mp4':
		case 'mkv':
		case 'avi':
		case 'mov':
		case 'webm':
		case 'flv':
			return <VideoIcon color="error" />;

		case 'mp3':
		case 'wav':
		case 'ogg':
		case 'flac':
		case 'm4a':
		case 'aac':
			return <AudioIcon color="secondary" />;

		case 'db':
		case 'sqlite':
		case 'sqlite3':
		case 'mdb':
		case 'rdb':
			return <DbIcon color="primary" />;

		default:
			return <GenericDocIcon color="action" />;
	}
};
