import { useState } from 'react';
import { Button, Card, Table, Tag, Typography, Space, Popconfirm, Message, Select } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Domain { id: number; name: string; }
interface DNSRecord {
  id: number; host: string; type: string; value: string;
  ttl: number; line: string; is_external: boolean;
}

export default function DNSRecordsPage() {
  const qc = useQueryClient();
  const [domainId, setDomainId] = useState<number>(0);

  const { data: domainsData } = useQuery({
    queryKey: ['domains-select'],
    queryFn: () => client.get('/domains', { params: { limit: 200 } }).then(r => r.data),
  });

  const { data, isLoading } = useQuery({
    queryKey: ['records', domainId],
    queryFn: () => client.get(`/domains/${domainId}/records`).then(r => r.data),
    enabled: domainId > 0,
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/domains/${domainId}/records/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['records', domainId] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '主机记录', dataIndex: 'host' },
    { title: '类型', dataIndex: 'type', width: 80 },
    { title: '值', dataIndex: 'value' },
    { title: 'TTL', dataIndex: 'ttl', width: 80 },
    { title: '线路', dataIndex: 'line', width: 100 },
    {
      title: '外部', dataIndex: 'is_external',
      render: (v: boolean) => <Tag color={v ? 'orange' : 'gray'}>{v ? '外部' : '托管'}</Tag>,
    },
    {
      title: '操作', render: (_: unknown, row: DNSRecord) => (
        <Popconfirm title="确认删除此记录？" onOk={() => deleteMut.mutate(row.id)}>
          <Button size="mini" type="text" status="danger">删除</Button>
        </Popconfirm>
      ),
    },
  ];

  const domains: Domain[] = domainsData?.items ?? [];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>DNS 记录</Typography.Title>}
      extra={
        <Space>
          <Select
            placeholder="选择域名"
            style={{ width: 200 }}
            value={domainId || undefined}
            onChange={(v) => setDomainId(Number(v))}
          >
            {domains.map(d => (
              <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>
            ))}
          </Select>
          <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['records', domainId] })} />
        </Space>
      }
    >
      <Table
        loading={isLoading}
        columns={columns}
        data={data?.items ?? []}
        rowKey="id"
        pagination={false}
        noDataElement={domainId === 0 ? <span>请先选择域名</span> : undefined}
      />
    </Card>
  );
}
