import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Message, Modal, Form, Select, Switch,
} from '@arco-design/web-react';
import { IconRefresh, IconSend } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Domain { id: number; name: string; }

const statusColor: Record<string, string> = {
  pending: 'orange', running: 'blue', success: 'green', failed: 'red',
};

export default function PublishJobsPage() {
  const qc = useQueryClient();
  const [triggerVisible, setTriggerVisible] = useState(false);
  const [triggerForm] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['publish-jobs'],
    queryFn: () => client.get('/publish-jobs').then(r => r.data),
    refetchInterval: 15_000,
  });

  const { data: domainsData } = useQuery({
    queryKey: ['domains-select'],
    queryFn: () => client.get('/domains', { params: { limit: 200 } }).then(r => r.data),
  });

  const { data: recordsData, isLoading: recordsLoading } = useQuery({
    queryKey: ['records-for-publish', triggerForm.getFieldValue('domain_id')],
    queryFn: () => {
      const id = triggerForm.getFieldValue('domain_id');
      return id ? client.get(`/domains/${id}/records`).then(r => r.data) : null;
    },
    enabled: triggerVisible,
  });

  const triggerMut = useMutation({
    mutationFn: (body: { domain_id: number; dry_run: boolean }) =>
      client.post(`/domains/${body.domain_id}/publish`, {
        dry_run: body.dry_run,
        records: recordsData?.items ?? [],
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['publish-jobs'] });
      Message.success('发布任务已提交');
      setTriggerVisible(false);
      triggerForm.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '发布失败');
    },
  });

  const domains: Domain[] = domainsData?.items ?? [];

  const domainMap = new Map(domains.map(d => [d.id, d.name]));

  const handleTrigger = async () => {
    const vals = await triggerForm.validate();
    triggerMut.mutate({ domain_id: Number(vals.domain_id), dry_run: vals.dry_run ?? false });
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    {
      title: '域名', dataIndex: 'domain_id',
      render: (v: number) => domainMap.get(v) ?? `#${v}`,
    },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    {
      title: '试运行', dataIndex: 'dry_run',
      render: (v: boolean) => <Tag color={v ? 'orange' : 'gray'}>{v ? '是' : '否'}</Tag>,
    },
    { title: '记录数', dataIndex: 'records_count', width: 80 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>发布任务</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['publish-jobs'] })} />
            <Button type="primary" icon={<IconSend />} onClick={() => setTriggerVisible(true)}>发起发布</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="发起 DNS 发布"
        visible={triggerVisible}
        onOk={handleTrigger}
        confirmLoading={triggerMut.isPending || recordsLoading}
        onCancel={() => { setTriggerVisible(false); triggerForm.resetFields(); }}
        unmountOnExit
      >
        <Form form={triggerForm} layout="vertical">
          <Form.Item label="选择域名" field="domain_id" rules={[{ required: true, message: '请选择域名' }]}>
            <Select
              placeholder="选择要发布的域名"
              onChange={() => qc.invalidateQueries({ queryKey: ['records-for-publish', triggerForm.getFieldValue('domain_id')] })}
            >
              {domains.map(d => <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item label="试运行（不实际写入 DNS）" field="dry_run" triggerPropName="checked" initialValue={false}>
            <Switch checkedText="试运行" uncheckedText="正式发布" />
          </Form.Item>
          <div style={{ color: '#86909c', fontSize: 12 }}>
            将把域名当前所有 DNS 记录发布到 CDN 服务商
          </div>
        </Form>
      </Modal>
    </>
  );
}
