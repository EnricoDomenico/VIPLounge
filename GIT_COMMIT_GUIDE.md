# 🚀 GUIA DE COMMIT E PUSH SEGURO

## ⚠️ ANTES DE FAZER `git push` - LEIA ISTO!

---

## 1. ✅ VERIFICAÇÕES FINAIS

### Verificar se .env está seguro (NÃO será enviado)

```bash
# Confirmar que .env está no .gitignore
cat .gitignore | grep ".env"
# Deve retornar: .env

# Listar arquivos que SERÃO commitados
git status

# ❌ Se .env aparecer como "modified" ou "new", PARE AQUI!
# Execute: git rm --cached .env
```

### Verificar se não há tokens no código

```bash
# Procurar por tokens hardcoded
grep -r "74539367-69b7-432a-934f-8d9050bade0c" .
grep -r "d769811d-2d05-4640-b756-b2bae62318cd" .

# Se retornar algo, é um problema!
```

---

## 2. 📋 ARQUIVOS QUE DEVEM SER COMMITADOS

```bash
# SEGUROS (sem dados sensíveis):
✅ cmd/
✅ internal/
✅ web/
✅ docs/
✅ config.yaml (sem tokens)
✅ .env.example (exemplo sem tokens)
✅ .gitignore (atualizado)
✅ go.mod, go.sum
✅ README.md
✅ SECURITY_AUDIT.md
✅ SECURITY_ANALYSIS_REPORT.md
✅ scripts/
✅ Dockerfile
✅ cloudbuild.yaml
```

```bash
# NÃO DEVEM SER COMMITADOS:
❌ .env (com tokens reais)
❌ bin/ (build artifacts)
❌ vendor/ (dependências)
❌ *.exe (executáveis)
❌ .vscode/ (configurações locais)
❌ .idea/ (IDE config)
```

---

## 3. 🔐 PASSO A PASSO - COMMIT SEGURO

### Passo 1: Verificar antes de adicionar

```bash
# Ver o que vai ser adicionado
git diff --cached

# Se vir tokens ou .env com dados reais, PARE!
git reset HEAD
```

### Passo 2: Adicionar arquivos com segurança

```bash
# Opção 1: Adicionar todos os arquivos seguros
git add .

# Opção 2: Adicionar seletivamente
git add cmd/ internal/ web/ *.md *.yaml *.mod scripts/ Dockerfile

# ❌ NUNCA fazer:
# git add -A  (pode incluir .env se não estiver bem no .gitignore)
# git add .env  (NUNCA!)
```

### Passo 3: Revisar tudo antes de commitar

```bash
# Ver exatamente o que será commitado
git status
git diff --cached

# Se tudo OK, commitar
git commit -m "Refactor: Adicionar segurança e documentação para produção"
```

### Passo 4: Verificação final antes de push

```bash
# Ver os commits que serão feitos push
git log origin/main..HEAD

# Verificar se tem algo estranho
# Se tiver token em algum commit, você pode fazer:
# git reset HEAD~1  (desfazer último commit)
```

### Passo 5: Push seguro

```bash
# Push para seu branch primeiro (não main)
git push origin feature/security-improvements

# Depois fazer Pull Request para revisar
```

---

## 4. 🆘 SE VOCÊ ACIDENTALMENTE COMMITOU UM TOKEN

### Opção A: Token commitou mas ainda não fez push

```bash
# Desfazer último commit (mantém alterações)
git reset --soft HEAD~1

# Editar .env para remover token
# Adicionar novamente sem o token
git add -p  # Adiciona seletivamente por pedaço
git commit -m "Remove: Remover arquivo .env do commit"

# Agora fazer push
```

### Opção B: Token já foi feito push 🔴

```bash
# ❌ IMEDIATO: Regenerar token em Superlogica/Rede Parcerias
# O token está comprometido!

# Remover do histórico Git
git filter-branch --tree-filter 'rm -f .env' -f

# Push forçado
git push -f origin main

# ⚠️ Isso reescreve o histórico!
# Avisar ao time para fazer rebase
```

---

## 5. ✅ CHECKLIST FINAL ANTES DE PUSH

```bash
# Executar todas as verificações

# 1. Security check
bash scripts/security-check.sh

# 2. Nenhum token no histórico
git log --all -p | grep -i "token\|secret\|password" | wc -l
# Deve retornar: 0

# 3. Nenhum token nos arquivos
grep -r "74539367-69b7-432a-934f-8d9050bade0c" --exclude-dir=.git .
# Não deve retornar nada

# 4. .env não está sendo tracked
git ls-files | grep ".env"
# Não deve retornar nada (apenas .env.example é OK)

# 5. Revisar o que vai subir
git status
git diff --cached | head -100

# Se tudo OK, fazer push
git push origin feature/security-improvements
```

---

## 6. 📝 TEMPLATE DE COMMIT MESSAGE

```bash
git commit -m "Security: Implementar headers HTTP e preparar para produção

- Add security middleware com headers (X-Frame-Options, CSP, HSTS)
- Add Firestore security rules template
- Add production setup scripts
- Add security audit report
- Never commit .env com tokens reais - usar Secret Manager
- Remove credenciais expostas do código"
```

---

## 7. 🚀 QUANDO TUDO ESTÁ PRONTO PARA SUBIR

```bash
# Verificação final
echo "🔐 Verificações de segurança:"
echo "✅ .env não será commitado"
echo "✅ Nenhum token no código"
echo "✅ Security headers implementados"
echo "✅ Firestore rules criadas"
echo "✅ Production scripts criados"
echo ""
echo "Pronto para fazer push!"

# Fazer push
git push origin feature/security-improvements

# Criar Pull Request no GitHub
# Descrever mudanças de segurança
# Aguardar aprovação
# Fazer merge
```

---

## 🛑 PARE SE:

```
❌ .env aparecer em git status como "modified"
❌ Ver tokens em git diff --cached
❌ Ver ".env" em git ls-files
❌ Ver QUALQUER credencial em git log
```

---

## ✅ TUDO PRONTO QUANDO:

```
✅ git status mostra apenas arquivos seguros
✅ git diff --cached tem apenas código seguro
✅ Nenhum .env em git ls-files
✅ Nenhum token em grep dos logs
✅ scripts/security-check.sh passou
✅ Revisor aprovou a Pull Request
```

---

## 🎯 WORKFLOW RECOMENDADO

```bash
# 1. Criar branch
git checkout -b feature/security-improvements

# 2. Fazer todas as alterações
# ... editar arquivos ...

# 3. Verificações
bash scripts/security-check.sh

# 4. Adicionar mudanças
git add -p  # Adiciona seletivamente

# 5. Revisar antes de commitar
git diff --cached

# 6. Commitar
git commit -m "Security: ..."

# 7. Verificação final
git log -1 -p

# 8. Push
git push origin feature/security-improvements

# 9. Criar Pull Request
# - Descrever mudanças
# - Listar verificações de segurança
# - Aguardar review

# 10. Merge após aprovação
```

---

## ❓ FAQ

**P: Posso fazer commit com .env?**  
R: ❌ **NUNCA!** Mesmo que tenha token fake, o .gitignore deve prevenir isso.

**P: Posso fazer push de um commit com token por acidente?**  
R: Se não fez push ainda: `git reset HEAD~1`  
   Se já fez push: Regenerar token imediatamente! Ele está comprometido!

**P: Como verifico se um token vazou?**  
R: `git log --all -p | grep "seu-token"`

**P: Preciso de um token para desenvolvimento?**  
R: Sim! Coloque em `.env` localmente. Ele não será commitado (está no .gitignore).

**P: Como usar tokens em produção se não commitam?**  
R: Usar Google Cloud Secret Manager ou variáveis de ambiente do Cloud Run.

---

## 📞 DÚVIDAS?

Consulte: [SECURITY_ANALYSIS_REPORT.md](SECURITY_ANALYSIS_REPORT.md)

---

**🔒 Lembre-se: Uma credencial vazada = dados de usuários em risco!**
