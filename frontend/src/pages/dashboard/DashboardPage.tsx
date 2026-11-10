import { Card, Grid, Statistic, Typography, Spin } from '@arco-design/web-react';
import { useQuery } from '@tanstack/react-query';
import client from '../../api/client';

const { Row, Col } = Grid;

interface Overview {
  domain_count: number;
  node_count: number;
  certificate_count: number;
  firing_alert_count: number;
}

export default function DashboardPage() {
  const { data, isLoading } = useQuery<Overview>({
    queryKey: ['dashboard-overview'],
    queryFn: () => client.get('/dashboard/overview').then((r) => r.data),
    refetchInterval: 30_000,
  });

  if (isLoading) return <Spin style={{ display: 'block', margin: '80px auto' }} />;

  return (
    <div>
      <Typography.Title heading={5} style={{ marginBottom: 24 }}>仪表盘</Typography.Title>

      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card>
            <Statistic title="域名总数" value={data?.domain_count ?? 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="CDN 节点" value={data?.node_count ?? 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="SSL 证书" value={data?.certificate_count ?? 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃告警"
              value={data?.firing_alert_count ?? 0}
              style={{ color: (data?.firing_alert_count ?? 0) > 0 ? 'red' : 'green' }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
}
