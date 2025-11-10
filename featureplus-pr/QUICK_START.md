# FeaturePlus CLI - Quick Start Guide

## ✅ Setup Complete!

The CLI has been built and added to your Windows PATH. 

---

## 🔄 Next Steps

### 1. **Restart Your Terminal**

**IMPORTANT:** You must close and reopen your terminal (PowerShell/CMD) for the PATH changes to take effect.

### 2. **Verify Installation**

Open a **new** terminal window and run:

```powershell
featureplus-pr --version
```

You should see:
```
FeaturePlus CLI v0.1.0
```

### 3. **Test from Any Directory**

Navigate to any folder and try:

```powershell
# Go to your home directory
cd ~

# Run the CLI
featureplus-pr --help
```

---

## 🚀 Usage

### Initialize a Project

```bash
# Navigate to your project directory
cd C:\path\to\your\project

# Initialize FeaturePlus
featureplus-pr init
```

### Connect to FeaturePlus

```bash
# Connect to a project (replace 1 with your project ID)
featureplus-pr connect 1
```

### Check Status

```bash
featureplus-pr status
```

---

## 📍 What Was Done

1. ✅ Built `featureplus-pr.exe` in: `d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr\`
2. ✅ Added to User PATH: `d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr`
3. ✅ CLI is now globally accessible

---

## 🔧 If It Doesn't Work

### Problem: "featureplus-pr is not recognized"

**Solution 1:** Restart your terminal
- Close all PowerShell/CMD windows
- Open a new one
- Try again

**Solution 2:** Verify PATH manually
1. Press `Win + X` → System
2. Advanced system settings → Environment Variables
3. Check User PATH contains: `d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr`

**Solution 3:** Run setup again
```powershell
cd d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr
.\setup-windows.ps1
```

---

## 📚 Available Commands

| Command | Description |
|---------|-------------|
| `featureplus-pr init` | Initialize FeaturePlus in current directory |
| `featureplus-pr connect <id>` | Connect directory to a project |
| `featureplus-pr status` | Show connection status |
| `featureplus-pr upload` | Upload PR info to FeaturePlus |
| `featureplus-pr list` | List pull requests |
| `featureplus-pr --help` | Show all commands |

---

## 🎯 Example Workflow

```powershell
# 1. Navigate to your project
cd C:\Users\YourName\projects\my-app

# 2. Initialize FeaturePlus
featureplus-pr init

# 3. Connect to project ID 1
featureplus-pr connect 1

# 4. Check status
featureplus-pr status

# Output:
# ╔════════════════════════════════════════════════╗
# ║         FeaturePlus Connection Status         ║
# ╚════════════════════════════════════════════════╝
#
# 📦 Project:       My App (1)
# 🌐 Server:        http://localhost:8080
# 📂 Path:          C:\Users\YourName\projects\my-app
# 🔗 Connected:     yes
# 🕓 Linked At:     2025-11-09 23:30:00
```

---

## ✨ You're All Set!

The FeaturePlus CLI is now installed and ready to use from any directory on your Windows machine, just like `git`, `node`, or `npm`.

**Remember:** Always restart your terminal after installation!
