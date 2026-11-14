import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Tooltip, Message, Modal, Form, Input, Select, Switch, Popconfirm,
  Drawer, Empty, Spin,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconDriveFile, IconSend } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Certificate {
  id: number;
  domain_name: string;
  status: string;
  acme_provider: string;
  not_after: string | null;
  auto_renew: boolean;
}
interface Domain { id: number; name: string; }
interface DNSAccount { id: number; name: string; provider_type: string; }

interface ProviderField {
  key: string; label: string; type: string;
  required?: boolean; placeholder?: string; help_text?: string;
  options?: { label: string; value: string }[];
}
interface ProviderDef {
  type: string; label: string; category: string; available: boolean;
  description?: string; fields: ProviderField[];
}

interface DeployJob {
  id: number; cert_id: number; name: string; deploy_type: string;
  status: string; last_deployed_at: string | null; last_error: string; created_at: string;
  credential_id?: number;
}

function daysUntil(dateStr: string | null): number {
  if (!dateStr) return -1;
  return Math.floor((new Date(dateStr).getTime() - Date.now()) / 86_400_000);
}
function expiryColor(days: number): string {
  if (days < 0) return 'red';
  if (days < 7) return 'red';
  if (days < 30) return 'orange';
  return 'green';
}

export default function CertificatesPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [form] = Form.useForm();
  const [, setSelectedDomainId] = useState<number>(0);

  // Deploy jobs drawer
  const [deployDrawerCert, setDeployDrawerCert] = useState<Certificate | null>(null);
  const [createJobVisible, setCreateJobVisible] = useState(false);
  const [jobForm] = Form.useForm();
  const [selectedDeployType, setSelectedDeployType] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['certificates'],
    queryFn: () => client.get('/certificates', { params: { limit: 50 } }).then((r) => r.data),
  });
  const { data: domainsData } = useQuery({
    queryKey: ['domains-select'],
    queryFn: () => client.get('/domains', { params: { limit: 200 } }).then(r => r.data),
  });
  const { data: dnsAccountsData } = useQuery({
    queryKey: ['dns-accounts'],
    queryFn: () => client.get('/dns-accounts').then(r => r.data),
  });
  const { data: providersData } = useQuery({
    queryKey: ['ssl-providers'],
    queryFn: () => client.get('/ssl/providers').then(r => r.data),
  });
  const { data: credentialsData } = useQuery({
    queryKey: ['ssl-credentials'],
    queryFn: () => client.get('/ssl/credentials').then(r => r.data),
  });
  const { data: deployJobsData, isLoading: jobsLoading } = useQuery({
    queryKey: ['deploy-jobs', deployDrawerCert?.id],
    queryFn: () => deployDrawerCert
      ? client.get(`/ssl/certs/${deployDrawerCert.id}/deploy-jobs`).then(r => r.data)
      : Promise.resolve({ items: [] }),
    enabled: !!deployDrawerCert,
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/certificates', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['certificates'] });
      Message.success('证书申请已提交');
      setVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '申请失败');
    },
  });
  const retriggerMut = useMutation({
    mutationFn: (id: number) => client.post(`/certificates/${id}/retrigger`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['certificates'] }); Message.success('已触发重签'); },
    onError: () => Message.error('操作失败'),
  });
  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/certificates/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['certificates'] }); Message.success('证书已删除'); },
    onError: () => Message.error('删除失败'),
  });
  const createJobMut = useMutation({
    mutationFn: (body: unknown) => client.post(`/ssl/certs/${deployDrawerCert!.id}/deploy-jobs`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['deploy-jobs', deployDrawerCert?.id] });
      Message.success('部署任务已创建');
      setCreateJobVisible(false);
      jobForm.resetFields();
      setSelectedDeployType('');
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });
  const executeJobMut = useMutation({
    mutationFn: (id: number) => client.post(`/ssl/deploy-jobs/${id}/execute`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['deploy-jobs', deployDrawerCert?.id] });
      Message.success('部署已执行');
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '执行失败');
    },
  });
  const deleteJobMut = useMutation({
    mutationFn: (id: number) => client.delete(`/ssl/deploy-jobs/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['deploy-jobs', deployDrawerCert?.id] }); Message.success('已删除'); },
  });

  const domains: Domain[] = domainsData?.items ?? [];
  const dnsAccounts: DNSAccount[] = dnsAccountsData?.items ?? [];
  const providers: ProviderDef[] = providersData?.items ?? [];
  const deployProviders = providers.filter(p => p.category === 'deploy' && p.available);
  const credentials: { id: number; name: string; provider_type: string; category: string }[] = credentialsData?.items ?? [];
  const deployJobs: DeployJob[] = deployJobsData?.items ?? [];

  const selectedDeployDef = deployProviders.find(p => p.type === selectedDeployType);

  const handleCreate = async () => {
    const vals = await form.validate();
    const selectedDomain = domains.find(d => d.id === vals.domain_id);
    createMut.mutate({
      ...vals,
      domain_name: vals.domain_name || selectedDomain?.name || '',
      acme_provider: vals.acme_provider || 'letsencrypt',
      challenge: vals.challenge || 'dns-01',
    });
  };

  const handleCreateJob = async () => {
    const vals = await jobForm.validate();
    const { name, deploy_type, credential_id, ...configFields } = vals;
    createJobMut.mutate({
      name,
      deploy_type,
      credential_id: credential_id || undefined,
      config_json: JSON.stringify(configFields),
    });
  };

  const statusColor: Record<string, string> = {
    active: 'green', pending: 'orange', issuing: 'blue',
    failed: 'red', expired: 'red', revoked: 'gray',
  };
  const jobStatusColor: Record<string, string> = {
    success: 'green', pending: 'gray', failed: 'red',
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '域名', dataIndex: 'domain_name' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    { title: 'ACME 供应商', dataIndex: 'acme_provider' },
    {
      title: '到期', dataIndex: 'not_after',
      render: (v: string | null) => {
        const days = daysUntil(v);
        if (days < 0) return <Tag color="gray">未知</Tag>;
        return (
          <Tooltip content={v ?? ''}>
            <Tag color={expiryColor(days)}>{days} 天</Tag>
          </Tooltip>
        );
      },
    },
    {
      title: '自动续签', dataIndex: 'auto_renew',
      render: (v: boolean) => <Tag color={v ? 'arcoblue' : 'gray'}>{v ? '是' : '否'}</Tag>,
    },
    {
      title: '操作', render: (_: unknown, row: Certificate) => (
        <Space>
          <Button size="mini" icon={<IconDriveFile />} onClick={() => setDeployDrawerCert(row)}>部署</Button>
          <Button size="mini" onClick={() => retriggerMut.mutate(row.id)}>重签</Button>
          <Popconfirm title="确认删除此证书？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const jobColumns = [
    { title: '名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'deploy_type' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={jobStatusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    {
      title: '最后执行', dataIndex: 'last_deployed_at',
      render: (v: string | null) => v ? new Date(v).toLocaleString('zh-CN') : '—',
    },
    {
      title: '操作', render: (_: unknown, row: DeployJob) => (
        <Space>
          <Popconfirm title="执行此部署任务？" onOk={() => executeJobMut.mutate(row.id)}>
            <Button size="mini" type="primary" icon={<IconSend />}>执行</Button>
          </Popconfirm>
          <Popconfirm title="确认删除？" onOk={() => deleteJobMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>SSL 证书</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['certificates'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>申请证书</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      {/* Deploy jobs drawer */}
      <Drawer
        title={`部署任务 — ${deployDrawerCert?.domain_name ?? ''}`}
        visible={!!deployDrawerCert}
        width={640}
        onCancel={() => setDeployDrawerCert(null)}
        footer={null}
      >
        <div style={{ marginBottom: 12, display: 'flex', justifyContent: 'flex-end' }}>
          <Button type="primary" icon={<IconPlus />} onClick={() => setCreateJobVisible(true)}>添加部署任务</Button>
        </div>
        {jobsLoading ? <Spin style={{ display: 'block', margin: '40px auto' }} /> : (
          deployJobs.length === 0
            ? <Empty description="暂无部署任务" />
            : <Table columns={jobColumns} data={deployJobs} rowKey="id" pagination={false} />
        )}
        {deployJobs.some(j => j.last_error) && (
          deployJobs.filter(j => j.last_error).map(j => (
            <div key={j.id} style={{ marginTop: 8, color: '#f53f3f', fontSize: 12 }}>
              [{j.name}] {j.last_error}
            </div>
          ))
        )}
      </Drawer>

      {/* Create deploy job modal */}
      <Modal
        title="添加部署任务"
        visible={createJobVisible}
        onOk={handleCreateJob}
        confirmLoading={createJobMut.isPending}
        onCancel={() => { setCreateJobVisible(false); jobForm.resetFields(); setSelectedDeployType(''); }}
        unmountOnExit
        style={{ width: 520 }}
      >
        <Form form={jobForm} layout="vertical">
          <Form.Item label="任务名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：部署到主站 Nginx" />
          </Form.Item>
          <Form.Item label="部署类型" field="deploy_type" rules={[{ required: true, message: '请选择类型' }]}>
            <Select placeholder="选择部署方式" onChange={v => setSelectedDeployType(v as string)}>
              {deployProviders.map(p => (
                <Select.Option key={p.type} value={p.type}>{p.label}</Select.Option>
              ))}
            </Select>
          </Form.Item>
          {selectedDeployDef?.fields.map(f => (
            <Form.Item
              key={f.key}
              label={f.label}
              field={f.key}
              rules={f.required ? [{ required: true, message: `请输入 ${f.label}` }] : []}
              extra={f.help_text ? <span style={{ fontSize: 11, color: '#86909c' }}>{f.help_text}</span> : undefined}
            >
              {f.type === 'textarea'
                ? <Input.TextArea rows={3} placeholder={f.placeholder} />
                : f.type === 'password'
                  ? <Input.Password placeholder={f.placeholder} />
                  : <Input placeholder={f.placeholder} />
              }
            </Form.Item>
          ))}
          {/* Optional: link a credential from the credential center */}
          {selectedDeployType && (
            <Form.Item label="关联凭据（可选）" field="credential_id"
              extra={<span style={{ fontSize: 11, color: '#86909c' }}>选择凭据中心中已存储的部署凭据</span>}
            >
              <Select allowClear placeholder="无（直接填写上方配置）">
                {credentials.filter(c => c.category === 'deploy').map(c => (
                  <Select.Option key={c.id} value={c.id}>{c.name} ({c.provider_type})</Select.Option>
                ))}
              </Select>
            </Form.Item>
          )}
        </Form>
      </Modal>

      {/* Create certificate modal */}
      <Modal
        title="申请 SSL 证书"
        visible={visible}
        onOk={handleCreate}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); setSelectedDomainId(0); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="域名" field="domain_id" rules={[{ required: true, message: '请选择域名' }]}>
            <Select
              placeholder="选择域名"
              onChange={v => {
                setSelectedDomainId(Number(v));
                const d = domains.find(x => x.id === Number(v));
                if (d) form.setFieldValue('domain_name', d.name);
              }}
            >
              {domains.map(d => <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item label="域名（可手动输入）" field="domain_name" rules={[{ required: true, message: '请输入域名' }]}>
            <Input placeholder="example.com" />
          </Form.Item>
          <Form.Item label="联系邮箱" field="account_email" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
            <Input placeholder="admin@example.com" />
          </Form.Item>
          <Form.Item label="ACME 供应商" field="acme_provider" initialValue="letsencrypt">
            <Select>
              <Select.Option value="letsencrypt">Let's Encrypt</Select.Option>
              <Select.Option value="zerossl">ZeroSSL</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="验证方式" field="challenge" initialValue="dns-01">
            <Select>
              <Select.Option value="dns-01">DNS-01</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="DNS 账户（DNS-01 验证用）" field="dns_account_id">
            <Select placeholder="选择 DNS 账户" allowClear>
              {dnsAccounts.map(a => (
                <Select.Option key={a.id} value={Number(a.id)}>{a.name} ({a.provider_type})</Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item label="自动续签" field="auto_renew" triggerPropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
          <Form.Item label="使用测试环境（Staging）" field="staging" triggerPropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
