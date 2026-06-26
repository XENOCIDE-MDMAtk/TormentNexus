import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const srcDir = path.resolve(__dirname, 'out');
const destDir = path.resolve(__dirname, '../../go/cmd/tormentnexus-gui/frontend/dist');

function copyRecursiveSync(src, dest) {
  const exists = fs.existsSync(src);
  const stats = exists && fs.statSync(src);
  const isDirectory = exists && stats.isDirectory();
  if (isDirectory) {
    if (!fs.existsSync(dest)) {
      fs.mkdirSync(dest, { recursive: true });
    }
    fs.readdirSync(src).forEach((childItemName) => {
      copyRecursiveSync(path.join(src, childItemName), path.join(dest, childItemName));
    });
  } else {
    fs.mkdirSync(path.dirname(dest), { recursive: true });
    fs.copyFileSync(src, dest);
  }
}

if (fs.existsSync(srcDir)) {
  console.log(`Copying Next.js export assets from ${srcDir} to ${destDir}...`);
  // Ensure the target dir exists and clean it up first
  if (fs.existsSync(destDir)) {
    fs.rmSync(destDir, { recursive: true, force: true });
  }
  copyRecursiveSync(srcDir, destDir);
  // Re-create the placeholder.txt so it doesn't break future builds
  fs.writeFileSync(path.join(destDir, 'placeholder.txt'), 'placeholder');
  console.log('Frontend assets copy complete!');
} else {
  console.error(`Source directory ${srcDir} does not exist. Run next build first.`);
}
