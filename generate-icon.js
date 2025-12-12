const sharp = require('sharp');
const toIco = require('to-ico');
const fs = require('fs');
const path = require('path');

const svgPath = path.join(__dirname, 'build', 'icon.svg');
const pngPath = path.join(__dirname, 'build', 'appicon.png');
const icoPath = path.join(__dirname, 'build', 'windows', 'icon.ico');
const trayIcoPath = path.join(__dirname, 'internal', 'tray', 'icon.ico');

async function convertIcon() {
    try {
        // 读取 SVG
        const svgBuffer = fs.readFileSync(svgPath);
        
        // 转换为 1024x1024 PNG
        await sharp(svgBuffer)
            .resize(1024, 1024)
            .png()
            .toFile(pngPath);
        
        console.log('✓ PNG 图标已生成:', pngPath);
        
        // 生成多尺寸的 PNG Buffer 用于 ICO
        const sizes = [256, 128, 64, 48, 32, 16];
        const pngBuffers = await Promise.all(
            sizes.map(size => 
                sharp(svgBuffer)
                    .resize(size, size)
                    .png()
                    .toBuffer()
            )
        );
        
        // 转换为 ICO
        const icoBuffer = await toIco(pngBuffers);
        fs.writeFileSync(icoPath, icoBuffer);
        fs.writeFileSync(trayIcoPath, icoBuffer);
        
        console.log('✓ ICO 图标已生成:', icoPath);
        console.log('✓ 托盘图标已生成:', trayIcoPath);
        console.log('\n图标生成完成！请运行 wails build 重新构建应用。');
        
    } catch (error) {
        console.error('错误:', error.message);
    }
}

convertIcon();
