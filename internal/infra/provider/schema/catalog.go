package schema

// ProviderField 描述凭据表单中的一个输入字段。
type ProviderField struct {
	Key         string        `json:"key"`
	Label       string        `json:"label"`
	Type        string        `json:"type"`     // text | password | textarea | select
	Required    bool          `json:"required"`
	Placeholder string        `json:"placeholder,omitempty"`
	HelpText    string        `json:"help_text,omitempty"`
	Options     []FieldOption `json:"options,omitempty"`
}

// FieldOption 用于 select 类型字段的选项。
type FieldOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ProviderDef 描述一个 Provider 类型及其凭据 schema。
type ProviderDef struct {
	Type        string          `json:"type"`
	Label       string          `json:"label"`
	Category    string          `json:"category"` // dns | deploy | notify | acme_ca
	Description string          `json:"description,omitempty"`
	Fields      []ProviderField `json:"fields"`
}

// Catalog 返回所有已知 Provider 的定义。
func Catalog() []ProviderDef { return allProviders }

// CatalogByCategory 按类别过滤。
func CatalogByCategory(category string) []ProviderDef {
	var out []ProviderDef
	for _, p := range allProviders {
		if p.Category == category {
			out = append(out, p)
		}
	}
	return out
}

var allProviders = []ProviderDef{
	// ---- DNS Providers ----
	{
		Type: "aliyun_dns", Label: "阿里云 DNS", Category: "dns",
		Description: "阿里云云解析 DNS（万网）",
		Fields: []ProviderField{
			{Key: "access_key_id", Label: "Access Key ID", Type: "text", Required: true},
			{Key: "access_key_secret", Label: "Access Key Secret", Type: "password", Required: true},
		},
	},
	{
		Type: "dnspod", Label: "腾讯云 DNSPod", Category: "dns",
		Description: "腾讯云 DNSPod（含 dnspod.cn）",
		Fields: []ProviderField{
			{Key: "secret_id", Label: "SecretId", Type: "text", Required: true},
			{Key: "secret_key", Label: "SecretKey", Type: "password", Required: true},
		},
	},
	{
		Type: "cloudflare_dns", Label: "Cloudflare DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_token", Label: "API Token", Type: "password", Required: true,
				HelpText: "需要 Zone:DNS:Edit 权限"},
		},
	},
	{
		Type: "huaweicloud_dns", Label: "华为云 DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "access_key_id", Label: "Access Key ID", Type: "text", Required: true},
			{Key: "secret_access_key", Label: "Secret Access Key", Type: "password", Required: true},
			{Key: "region", Label: "Region", Type: "text", Placeholder: "cn-north-4"},
		},
	},
	{
		Type: "aws_route53", Label: "AWS Route53", Category: "dns",
		Fields: []ProviderField{
			{Key: "access_key_id", Label: "AWS Access Key ID", Type: "text", Required: true},
			{Key: "secret_access_key", Label: "AWS Secret Access Key", Type: "password", Required: true},
			{Key: "region", Label: "Region", Type: "text", Placeholder: "us-east-1"},
			{Key: "hosted_zone_id", Label: "Hosted Zone ID（可选）", Type: "text"},
		},
	},
	{
		Type: "godaddy", Label: "GoDaddy", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Type: "text", Required: true},
			{Key: "api_secret", Label: "API Secret", Type: "password", Required: true},
		},
	},
	{
		Type: "namecheap", Label: "Namecheap", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_user", Label: "API User", Type: "text", Required: true},
			{Key: "api_key", Label: "API Key", Type: "password", Required: true},
		},
	},
	{
		Type: "digitalocean_dns", Label: "DigitalOcean DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "auth_token", Label: "Personal Access Token", Type: "password", Required: true},
		},
	},
	{
		Type: "vultr_dns", Label: "Vultr DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true},
		},
	},
	{
		Type: "linode_dns", Label: "Linode / Akamai DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "token", Label: "Personal Access Token", Type: "password", Required: true},
		},
	},
	{
		Type: "hetzner_dns", Label: "Hetzner DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true},
		},
	},
	{
		Type: "dnsimple", Label: "DNSimple", Category: "dns",
		Fields: []ProviderField{
			{Key: "oauth_token", Label: "OAuth Token", Type: "password", Required: true},
		},
	},
	{
		Type: "ns1", Label: "NS1 / IBM NS1 Connect", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true},
		},
	},
	{
		Type: "porkbun", Label: "Porkbun", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Type: "text", Required: true},
			{Key: "secret_api_key", Label: "Secret API Key", Type: "password", Required: true},
		},
	},
	{
		Type: "gandi", Label: "Gandi", Category: "dns",
		Fields: []ProviderField{
			{Key: "personal_access_token", Label: "Personal Access Token", Type: "password", Required: true},
		},
	},
	{
		Type: "vercel_dns", Label: "Vercel DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "auth_token", Label: "Auth Token", Type: "password", Required: true},
		},
	},
	{
		Type: "duckdns", Label: "DuckDNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "token", Label: "Token", Type: "password", Required: true},
		},
	},
	{
		Type: "ovh_dns", Label: "OVH", Category: "dns",
		Fields: []ProviderField{
			{Key: "endpoint", Label: "Endpoint", Type: "text", Required: true, Placeholder: "ovh-eu"},
			{Key: "application_key", Label: "Application Key", Type: "text", Required: true},
			{Key: "application_secret", Label: "Application Secret", Type: "password", Required: true},
			{Key: "consumer_key", Label: "Consumer Key", Type: "text", Required: true},
		},
	},
	{
		Type: "azure_dns", Label: "Azure DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "client_id", Label: "Client ID", Type: "text", Required: true},
			{Key: "client_secret", Label: "Client Secret", Type: "password", Required: true},
			{Key: "subscription_id", Label: "Subscription ID", Type: "text", Required: true},
			{Key: "tenant_id", Label: "Tenant ID", Type: "text", Required: true},
			{Key: "resource_group", Label: "Resource Group", Type: "text", Required: true},
		},
	},
	{
		Type: "google_cloud_dns", Label: "Google Cloud DNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "service_account_json", Label: "Service Account JSON", Type: "textarea", Required: true,
				HelpText: "具有 DNS Administrator 权限的服务账号 JSON 密钥"},
		},
	},
	{
		Type: "namesilo", Label: "NameSilo", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true},
		},
	},
	{
		Type: "cloudns", Label: "ClouDNS", Category: "dns",
		Fields: []ProviderField{
			{Key: "auth_id", Label: "Auth ID", Type: "text", Required: true},
			{Key: "auth_password", Label: "Auth Password", Type: "password", Required: true},
		},
	},
	{
		Type: "dynu", Label: "Dynu", Category: "dns",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true},
		},
	},

	// ---- Deploy Providers ----
	{
		Type: "local_file", Label: "本地文件", Category: "deploy",
		Description: "将证书写入本地文件系统",
		Fields: []ProviderField{
			{Key: "cert_path", Label: "证书文件路径", Type: "text", Required: true,
				Placeholder: "/etc/ssl/certs/example.com.crt"},
			{Key: "key_path", Label: "私钥文件路径", Type: "text", Required: true,
				Placeholder: "/etc/ssl/private/example.com.key"},
		},
	},
	{
		Type: "nginx_ssh", Label: "Nginx (SSH)", Category: "deploy",
		Description: "通过 SSH 将证书上传至远程 Nginx 服务器并重载",
		Fields: []ProviderField{
			{Key: "host", Label: "SSH 主机", Type: "text", Required: true, Placeholder: "192.168.1.100"},
			{Key: "port", Label: "SSH 端口", Type: "text", Placeholder: "22"},
			{Key: "username", Label: "SSH 用户名", Type: "text", Required: true, Placeholder: "root"},
			{Key: "password", Label: "SSH 密码", Type: "password"},
			{Key: "private_key", Label: "SSH 私钥", Type: "textarea",
				HelpText: "PEM 格式的 SSH 私钥（密码或私钥二选一）"},
			{Key: "cert_path", Label: "证书路径", Type: "text", Required: true,
				Placeholder: "/etc/nginx/ssl/cert.pem"},
			{Key: "key_path", Label: "私钥路径", Type: "text", Required: true,
				Placeholder: "/etc/nginx/ssl/key.pem"},
			{Key: "reload_cmd", Label: "重载命令", Type: "text",
				Placeholder: "systemctl reload nginx"},
		},
	},
	{
		Type: "aliyun_esa_deploy", Label: "阿里云 ESA 证书绑定", Category: "deploy",
		Description: "将证书上传至阿里云 ESA 并绑定到站点",
		Fields: []ProviderField{
			{Key: "access_key_id", Label: "Access Key ID", Type: "text", Required: true},
			{Key: "access_key_secret", Label: "Access Key Secret", Type: "password", Required: true},
			{Key: "site_id", Label: "站点 ID", Type: "text", Required: true},
		},
	},
	{
		Type: "cloudflare_deploy", Label: "Cloudflare 自定义证书", Category: "deploy",
		Description: "将证书上传至 Cloudflare 自定义证书",
		Fields: []ProviderField{
			{Key: "api_token", Label: "API Token", Type: "password", Required: true,
				HelpText: "需要 Zone:SSL and Certificates:Edit 权限"},
			{Key: "zone_id", Label: "Zone ID", Type: "text", Required: true},
		},
	},
	{
		Type: "tencent_eo_deploy", Label: "腾讯云 EdgeOne 证书", Category: "deploy",
		Fields: []ProviderField{
			{Key: "secret_id", Label: "SecretId", Type: "text", Required: true},
			{Key: "secret_key", Label: "SecretKey", Type: "password", Required: true},
			{Key: "zone_id", Label: "Zone ID", Type: "text", Required: true},
		},
	},
	{
		Type: "ssh_generic", Label: "SSH 通用部署", Category: "deploy",
		Description: "通过 SSH 执行自定义部署脚本",
		Fields: []ProviderField{
			{Key: "host", Label: "SSH 主机", Type: "text", Required: true},
			{Key: "port", Label: "SSH 端口", Type: "text", Placeholder: "22"},
			{Key: "username", Label: "SSH 用户名", Type: "text", Required: true},
			{Key: "password", Label: "SSH 密码", Type: "password"},
			{Key: "private_key", Label: "SSH 私钥", Type: "textarea"},
			{Key: "pre_deploy_cmd", Label: "部署前命令", Type: "textarea"},
			{Key: "post_deploy_cmd", Label: "部署后命令", Type: "textarea"},
		},
	},

	// ---- ACME CA Providers ----
	{
		Type: "letsencrypt", Label: "Let's Encrypt", Category: "acme_ca",
		Description: "免费、自动化证书颁发机构（推荐）",
		Fields: []ProviderField{
			{Key: "email", Label: "账户邮箱", Type: "text", Required: true},
			{Key: "staging", Label: "环境", Type: "select",
				Options: []FieldOption{
					{Label: "生产环境", Value: "false"},
					{Label: "测试环境（不受信任）", Value: "true"},
				}},
		},
	},
	{
		Type: "zerossl", Label: "ZeroSSL", Category: "acme_ca",
		Description: "ZeroSSL ACME 接口，支持商业证书",
		Fields: []ProviderField{
			{Key: "email", Label: "账户邮箱", Type: "text", Required: true},
			{Key: "eab_kid", Label: "EAB KID（可选）", Type: "text",
				HelpText: "ZeroSSL External Account Binding"},
			{Key: "eab_hmac", Label: "EAB HMAC Key（可选）", Type: "password"},
		},
	},
	{
		Type: "buypass", Label: "Buypass Go SSL", Category: "acme_ca",
		Fields: []ProviderField{
			{Key: "email", Label: "账户邮箱", Type: "text", Required: true},
		},
	},

	// ---- Notification Providers ----
	{
		Type: "telegram_notify", Label: "Telegram", Category: "notify",
		Fields: []ProviderField{
			{Key: "bot_token", Label: "Bot Token", Type: "password", Required: true,
				HelpText: "从 @BotFather 获取"},
			{Key: "chat_id", Label: "Chat ID", Type: "text", Required: true},
		},
	},
	{
		Type: "dingtalk_notify", Label: "钉钉机器人", Category: "notify",
		Fields: []ProviderField{
			{Key: "webhook_url", Label: "Webhook URL", Type: "text", Required: true},
			{Key: "secret", Label: "加签密钥（可选）", Type: "password"},
		},
	},
	{
		Type: "feishu_notify", Label: "飞书机器人", Category: "notify",
		Fields: []ProviderField{
			{Key: "webhook_url", Label: "Webhook URL", Type: "text", Required: true},
		},
	},
	{
		Type: "slack_notify", Label: "Slack", Category: "notify",
		Fields: []ProviderField{
			{Key: "webhook_url", Label: "Incoming Webhook URL", Type: "text", Required: true},
		},
	},
	{
		Type: "smtp_notify", Label: "邮件 (SMTP)", Category: "notify",
		Fields: []ProviderField{
			{Key: "host", Label: "SMTP 主机", Type: "text", Required: true, Placeholder: "smtp.example.com"},
			{Key: "port", Label: "端口", Type: "text", Required: true, Placeholder: "465"},
			{Key: "username", Label: "用户名", Type: "text", Required: true},
			{Key: "password", Label: "密码", Type: "password", Required: true},
			{Key: "from", Label: "发件人", Type: "text", Required: true},
			{Key: "to", Label: "收件人（逗号分隔）", Type: "text", Required: true},
			{Key: "tls", Label: "加密方式", Type: "select",
				Options: []FieldOption{
					{Label: "SSL/TLS", Value: "ssl"},
					{Label: "STARTTLS", Value: "starttls"},
					{Label: "无加密", Value: "none"},
				}},
		},
	},
	{
		Type: "webhook_notify", Label: "Webhook", Category: "notify",
		Fields: []ProviderField{
			{Key: "url", Label: "Webhook URL", Type: "text", Required: true},
			{Key: "headers", Label: "自定义请求头（JSON）", Type: "textarea",
				Placeholder: `{"Authorization": "Bearer token"}`},
		},
	},
	{
		Type: "discord_notify", Label: "Discord", Category: "notify",
		Fields: []ProviderField{
			{Key: "webhook_url", Label: "Discord Webhook URL", Type: "text", Required: true},
		},
	},
	{
		Type: "pushover_notify", Label: "Pushover", Category: "notify",
		Fields: []ProviderField{
			{Key: "user_key", Label: "User Key", Type: "text", Required: true},
			{Key: "api_token", Label: "API Token", Type: "password", Required: true},
		},
	},
}
