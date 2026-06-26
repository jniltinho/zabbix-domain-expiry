# HOWTO — Configurar no Zabbix 7

Guia passo a passo para monitorar expiração de domínios com o binário `check_domain` no **Zabbix 7.x**.

> O template `zbx_domain_expiry.yaml` foi exportado no formato 6.4 e é compatível com Zabbix 7. Os itens do tipo **External check** são executados pelo **Zabbix Server** (ou pelo **Proxy**, se o host monitorado estiver atrás de um proxy).

## Índice

1. [Pré-requisitos](#1-pré-requisitos)
2. [Instalar o binário](#2-instalar-o-binário)
3. [Configurar o Zabbix Server](#3-configurar-o-zabbix-server)
4. [Importar o template](#4-importar-o-template)
5. [Criar o host do domínio](#5-criar-o-host-do-domínio)
6. [Testar a coleta](#6-testar-a-coleta)
7. [Ajustar macros e alertas](#7-ajustar-macros-e-alertas)
8. [Monitorar vários domínios](#8-monitorar-vários-domínios)
9. [Solução de problemas](#9-solução-de-problemas)

---

## 1. Pré-requisitos

| Item | Requisito |
|------|-----------|
| Zabbix | 7.0 ou superior |
| SO do servidor Zabbix | GNU/Linux **amd64** |
| Acesso | Shell no servidor Zabbix (ou no proxy responsável pelo host) |
| Rede | Saída HTTPS (RDAP) e TCP/43 (WHOIS) liberadas |

Não é necessário instalar `curl`, `jq`, `whois` nem outras dependências — o binário Go é autocontido.

---

## 2. Instalar o binário

### Opção A — Release (recomendado)

No **servidor Zabbix** (ou no proxy, se aplicável):

```bash
VERSION=0.0.1
curl -LO "https://github.com/jniltinho/zabbix-domain-expiry/releases/download/v${VERSION}/check_domain_${VERSION}_linux_amd64.tar.gz"
tar -xzf "check_domain_${VERSION}_linux_amd64.tar.gz"
sudo install -m 755 -o zabbix -g zabbix check_domain /usr/lib/zabbix/externalscripts/check_domain
```

### Opção B — Compilar no servidor

```bash
git clone https://github.com/jniltinho/zabbix-domain-expiry.git
cd zabbix-domain-expiry
make build-linux-amd64
sudo install -m 755 -o zabbix -g zabbix build/check_domain-linux-amd64 /usr/lib/zabbix/externalscripts/check_domain
```

### Validar manualmente

Execute como usuário `zabbix` para confirmar permissões e conectividade:

```bash
sudo -u zabbix /usr/lib/zabbix/externalscripts/check_domain -d example.com -r 0 -s 0 -w 30 -c 7
```

Saída esperada (JSON em stdout):

```json
{"state":"OK","days_left":365,"days_since_expired":0,"expire_date":"2026-08-13","message":"State: OK ; Days left: 365 ; Expire date: 2026-08-13"}
```

---

## 3. Configurar o Zabbix Server

### Diretório de external scripts

Confirme o caminho em `/etc/zabbix/zabbix_server.conf`:

```ini
### Option: ExternalScripts
ExternalScripts=/usr/lib/zabbix/externalscripts
```

Se alterar o caminho, reinicie o servidor:

```bash
sudo systemctl restart zabbix-server
```

### Zabbix Proxy

Se o host do domínio for monitorado por um **proxy**, o binário deve estar instalado no **proxy**, não apenas no servidor central. Configure `ExternalScripts` em `/etc/zabbix/zabbix_proxy.conf` e reinicie o proxy:

```bash
sudo systemctl restart zabbix-proxy
```

### Permissões

O processo do Zabbix executa os scripts como usuário `zabbix`. Garanta:

```bash
ls -l /usr/lib/zabbix/externalscripts/check_domain
# -rwxr-xr-x 1 zabbix zabbix ... check_domain
```

---

## 4. Importar o template

1. Acesse a interface web do Zabbix 7
2. Vá em **Data collection → Templates**
3. Clique em **Import** (canto superior direito)
4. Selecione o arquivo `zbx_domain_expiry.yaml`
5. Na tela de importação, marque:
   - **Templates** — Create new / Update existing (conforme necessário)
   - **Items**, **Triggers**, **Template groups**
6. Clique em **Import**

Após a importação, o template **Domain Expiry** deve aparecer em **Data collection → Templates**.

### Atualizar template existente

Para upgrade de versões anteriores (shell script `check_domain.sh`):

1. Substitua o binário em `externalscripts/`
2. Reimporte `zbx_domain_expiry.yaml` com **Update existing** habilitado
3. Confirme que a key do item master mudou de `check_domain.sh[...]` para `check_domain[...]`

---

## 5. Criar o host do domínio

O template usa `{HOST.HOST}` como nome do domínio. O **Host name** do host no Zabbix deve ser exatamente o domínio a monitorar.

1. Vá em **Data collection → Hosts**
2. Clique em **Create host**
3. Preencha:

| Campo | Exemplo | Observação |
|-------|---------|------------|
| **Host name** | `example.com` | Deve ser o domínio real |
| **Visible name** | `Example.com expiry` | Apenas exibição |
| **Host groups** | `Domains` (ou outro) | Opcional |
| **Interfaces** | — | Não é necessária interface para external check |

4. Na aba **Templates**, adicione o template **Domain Expiry**
5. Clique em **Add** / **Update**

> **Importante:** não use prefixos no Host name (ex.: `zabbix.example.com` em vez de `example.com`), a menos que esse seja realmente o domínio a consultar.

---

## 6. Testar a coleta

### Executar item manualmente

1. Abra o host criado → aba **Items**
2. Localize o item **Check Domain** (tipo External check)
3. Marque o item e clique em **Execute now**
4. Aguarde alguns segundos e verifique **Latest data**

### Itens esperados

| Item | Valor esperado |
|------|----------------|
| Check Domain | JSON completo |
| State | `OK`, `WARNING`, `CRITICAL` ou `UNKNOWN` |
| Days Left | Número de dias restantes |
| Expire Date | Data no formato `YYYY-MM-DD` |
| Message | Mensagem descritiva do check |

### Verificar triggers

Em **Monitoring → Problems**, confirme que não há alertas indevidos após a primeira coleta bem-sucedida.

---

## 7. Ajustar macros e alertas

As macros podem ser definidas no template ou sobrescritas por host.

Vá em **Data collection → Templates → Domain Expiry → Macros** (ou nas macros do host).

| Macro | Padrão | Descrição |
|-------|--------|-----------|
| `{$EXP_CRIT}` | `7` | Dias restantes para alerta **High** |
| `{$EXP_WARN}` | `30` | Dias restantes para alerta **Warning** |
| `{$RDAP_SERVER}` | *(vazio)* | URL do servidor RDAP; vazio usa bootstrap IANA |
| `{$WHOIS_SERVER}` | *(vazio)* | Servidor WHOIS; vazio usa lookup interno |

### Exemplo — TLD com RDAP customizado

Para um domínio `.uk`, se necessário:

```
{$RDAP_SERVER} = https://rdap.nominet.uk/uk-domain/
```

Deixe vazio para a maioria dos domínios — o binário resolve o servidor RDAP automaticamente.

### Triggers incluídas

| Prioridade | Condição |
|------------|----------|
| Not classified | Estado `UNKNOWN` (falha na consulta) |
| Disaster | Domínio expirado |
| High | Expira em ≤ `{$EXP_CRIT}` dias |
| Warning | Expira em ≤ `{$EXP_WARN}` dias |

---

## 8. Monitorar vários domínios

Cada domínio = **um host** no Zabbix.

```
Host: example.com     → template Domain Expiry
Host: meusite.com.br  → template Domain Expiry
Host: outro.org       → template Domain Expiry
```

O intervalo padrão de coleta é **1 dia** (`1d`), adequado para evitar rate limit de servidores WHOIS/RDAP. Para alterar:

1. Abra o item **Check Domain** no template ou no host
2. Ajuste **Update interval** (ex.: `12h`, `1d`)

---

## 9. Solução de problemas

### Item em estado NOT SUPPORTED

| Causa | Solução |
|-------|---------|
| Binário ausente | Instale em `/usr/lib/zabbix/externalscripts/check_domain` |
| Sem permissão de execução | `chmod 755` e owner `zabbix:zabbix` |
| Caminho `ExternalScripts` incorreto | Verifique `zabbix_server.conf` / `zabbix_proxy.conf` |
| Script no servidor errado | Instale no proxy se o host usa proxy |

### State = UNKNOWN

Teste manual com debug:

```bash
sudo -u zabbix /usr/lib/zabbix/externalscripts/check_domain -d example.com -r 0 -s 0 -z
```

Mensagens de debug vão para **stderr**; o JSON de resultado continua em stdout.

Causas comuns:

- Domínio inexistente ou sem dados de expiração pública
- Bloqueio de rede (firewall sem saída HTTPS/43)
- Rate limit do servidor WHOIS (aguarde ou aumente o intervalo de coleta)
- TLD com formato RDAP não padrão

### JSONPath não extrai valores

Confirme que o item **Check Domain** retorna JSON válido em **Latest data**. Os itens dependentes (`Days Left`, `State`, etc.) dependem desse master.

### Ver logs do Zabbix

```bash
# RHEL/Rocky
sudo tail -f /var/log/zabbix/zabbix_server.log

# Debian/Ubuntu
sudo tail -f /var/log/zabbix-server/zabbix_server.log
```

Procure por erros relacionados a `check_domain` ou `External script`.

### Testar conectividade RDAP/WHOIS

```bash
curl -s "https://rdap.org/domain/example.com" | head
whois example.com | head
```

---

## Referências

- [Zabbix 7 — External check](https://www.zabbix.com/documentation/7.0/en/manual/config/items/itemtypes/external)
- [Zabbix 7 — Template import](https://www.zabbix.com/documentation/7.0/en/manual/xml_export_import/templates)
- [Repositório do projeto](https://github.com/jniltinho/zabbix-domain-expiry)
- [README](README.md)