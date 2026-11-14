import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Tooltip, Message, Modal, Form, Input, Select, Switch,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
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

  const domains: Domain[] = domainsData?.items ?? [];
  const dnsAccounts: DNSAccount[] = dnsAccountsData?.items ?? [];

  const statusColor: Record<string, string> = {
    active: 'green', pending: 'orange', issuing: 'blue',
    failed: 'red', expired: 'red', revoked: 'gray',
  };

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
          <Button size="mini" onClick={() => retriggerMut.mutate(row.id)}>重签</Button>
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
