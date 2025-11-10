# FeaturePlus CLI - Windows Setup Guide

This guide will help you set up the FeaturePlus CLI so you can use it from any directory, just like `git` or `node`.

---

## 🚀 Quick Setup (Recommended)

### Option 1: Automated Setup with PowerShell

1. **Open PowerShell as Administrator** (Right-click → Run as Administrator)

2. **Navigate to the CLI directory:**
   ```powershell
   cd d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr
   ```

3. **Run the setup script:**
   ```powershell
   .\setup-windows.ps1
   ```

4. **Restart your terminal** (close and reopen PowerShell/CMD)

5. **Test it:**
   ```powershell
   featureplus-pr --version
   ```

**Done!** You can now use `featureplus-pr` from any directory.

---

## 📋 Manual Setup

If you prefer to set it up manually or the script doesn't work:

### Step 1: Build the Executable

```powershell
cd d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr
go build -o featureplus-pr.exe
```

### Step 2: Add to PATH

#### Method A: PowerShell Command (Easiest)

Run this command in PowerShell:

```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr", "User")
```

#### Method B: GUI (Windows Settings)

1. Press `Win + X` and select **System**
2. Click **Advanced system settings** (on the right)
3. Click **Environment Variables** button
4. Under **User variables**, select **Path** and click **Edit**
5. Click **New** and add:
   ```
   d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr
   ```
6. Click **OK** on all dialogs

#### Method C: Command Prompt (CMD)

```cmd
setx PATH "%PATH%;d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr"
```

### Step 3: Restart Terminal

**Important:** Close and reopen your terminal for PATH changes to take effect.

### Step 4: Verify

```powershell
featureplus-pr --version
```

You should see:
```
FeaturePlus CLI v0.1.0
```

---

## 🎯 Usage

Once installed, you can use these commands from **any directory**:

```bash
# Initialize FeaturePlus in current directory
featureplus-pr init

# Connect to a project
featureplus-pr connect 1

# Check connection status
featureplus-pr status

# See all commands
featureplus-pr --help
```

---

## 🔧 Troubleshooting

### "featureplus-pr is not recognized..."

**Problem:** The command is not found after adding to PATH.

**Solutions:**
1. **Restart your terminal** - PATH changes require a new terminal session
2. **Check PATH was added correctly:**
   ```powershell
   $env:Path -split ';' | Select-String "featureplus"
   ```
3. **Verify the executable exists:**
   ```powershell
   Test-Path "d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr\featureplus-pr.exe"
   ```

### PowerShell Execution Policy Error

**Problem:** "cannot be loaded because running scripts is disabled"

**Solution:** Run PowerShell as Administrator and execute:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Build Fails

**Problem:** `go build` command fails

**Solutions:**
1. **Ensure Go is installed:**
   ```powershell
   go version
   ```
2. **Check you're in the correct directory:**
   ```powershell
   cd d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr
   ```
3. **Run go mod tidy:**
   ```powershell
   go mod tidy
   go build -o featureplus-pr.exe
   ```

---

## 🎨 Alternative: Create an Alias

If you don't want to modify PATH, you can create a PowerShell alias:

1. **Edit your PowerShell profile:**
   ```powershell
   notepad $PROFILE
   ```

2. **Add this line:**
   ```powershell
   function featureplus-pr { & "d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr\featureplus-pr.exe" $args }
   ```

3. **Reload profile:**
   ```powershell
   . $PROFILE
   ```

---

## 📦 Uninstall

To remove the CLI from PATH:

### PowerShell:
```powershell
$path = [Environment]::GetEnvironmentVariable("Path", "User")
$newPath = ($path.Split(';') | Where-Object { $_ -ne "d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr" }) -join ';'
[Environment]::SetEnvironmentVariable("Path", $newPath, "User")
```

### GUI:
1. Open Environment Variables (Win + X → System → Advanced → Environment Variables)
2. Edit User PATH
3. Remove the FeaturePlus CLI entry
4. Click OK

---

## ✅ Verification Checklist

After setup, verify everything works:

- [ ] `featureplus-pr --version` shows version
- [ ] `featureplus-pr --help` shows help menu
- [ ] Can run from any directory (test in `C:\` or your home folder)
- [ ] `featureplus-pr init` creates `.featureplus` folder
- [ ] `featureplus-pr status` works

---

## 📞 Need Help?

If you encounter issues:

1. Check the troubleshooting section above
2. Verify Go is installed: `go version`
3. Ensure you restarted your terminal
4. Try the manual setup method if automated script fails
