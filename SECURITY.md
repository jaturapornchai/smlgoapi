# 🔒 Security Notice

## ⚠️ IMPORTANT: GitHub Copilot Restrictions

The following files and folders are **PROTECTED** from GitHub Copilot Agent modifications:

### 🚫 **DO NOT MODIFY:**
- `smlgoapi.json` - Contains sensitive database credentials
- `docs/` folder - Project documentation 
- Any `*.json` files in `config/` folder
- `.env` files

### 🔧 **How to Setup Configuration:**

1. **Copy the template:**
   ```bash
   cp smlgoapi.template.json smlgoapi.json
   ```

2. **Edit `smlgoapi.json` with your actual values:**
   - Database credentials
   - JWT secrets
   - Server configuration

3. **NEVER commit `smlgoapi.json` to Git!**

### 🛡️ **Protection Mechanisms:**

1. **`.copilotignore`** - Tells GitHub Copilot to ignore these files
2. **`.gitignore`** - Prevents committing sensitive files
3. **`.gitattributes`** - Marks files as binary/generated
4. **VS Code settings** - Workspace-level Copilot exclusions

### 🎯 **For Developers:**

- Use `smlgoapi.template.json` for reference
- Store sensitive values in environment variables in production
- Never share actual configuration files
- Always use HTTPS in production

### 📋 **Quick Commands:**

```bash
# Setup new environment
make setup

# Check configuration
make check-config

# Clean sensitive files
make clean
```

---
**Remember: Security is everyone's responsibility! 🔐**
