import React, { useState, useEffect } from 'react';
import { Layout, Table, Checkbox, Card, Typography, Space, Tag, theme, Divider, Empty, ConfigProvider, Statistic, Row, Col, Badge, Button, Modal, Form, InputNumber, Tooltip, Drawer, Timeline } from 'antd';
import { DatabaseOutlined, HddOutlined, SyncOutlined, SettingOutlined, DashboardOutlined, HistoryOutlined } from '@ant-design/icons';
import axios from 'axios';
import moment from 'moment';

const { Header, Content, Sider } = Layout;
const { Title, Text } = Typography;

// Configure axios base URL - Pointing to port 8085 as per config
const api = axios.create({
  baseURL: 'http://127.0.0.1:8085/api/v1',
});

const App = () => {
  const [devices, setDevices] = useState([]);
  const [selectedStationIDs, setSelectedStationIDs] = useState([]);
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [limit, setLimit] = useState(50);
  const [refreshInterval, setRefreshInterval] = useState(5);
  const [stats, setStats] = useState({ total: 0, online: 0, offline: 0 });
  const [hbStatus, setHbStatus] = useState({});
  const [hbConfig, setHbConfig] = useState({ check_interval: 60, default_timeout: 300 });
  const [isConfigModalOpen, setIsConfigModalOpen] = useState(false);
  const [isTrendDrawerOpen, setIsTrendDrawerOpen] = useState(false);
  const [trendData, setTrendData] = useState([]);
  const [configLoading, setConfigLoading] = useState(false);
  const [form] = Form.useForm();

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // Fetch metadata and stats
  useEffect(() => {
    fetchDevices();
    fetchStats();
    fetchHbStatus();
    fetchHbConfig();
    const interval = setInterval(() => {
      fetchDevices();
      fetchStats();
      fetchHbStatus();
    }, 10000);
    return () => clearInterval(interval);
  }, []);

  // Poll data based on selection
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, refreshInterval * 1000);
    return () => clearInterval(interval);
  }, [selectedStationIDs, limit, refreshInterval]);

  const fetchDevices = async () => {
    try {
      const res = await api.get('/devices');
      if (res.data.code === 0) {
        const devList = res.data.data || [];
        setDevices(devList);
        // Auto-select all if selection is empty (first load)
        if (selectedStationIDs.length === 0 && devList.length > 0) {
          setSelectedStationIDs(devList.map(d => d.id));
        }
      }
    } catch (error) {
      console.error('Failed to fetch devices:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const res = await api.get('/statistics/online');
      if (res.data.code === 0) {
        setStats(res.data.data);
      }
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  const fetchHbStatus = async () => {
    try {
      const res = await api.get('/heartbeat/status');
      if (res.data.code === 0) {
        setHbStatus(res.data.data || {});
      }
    } catch (error) {
      console.error('Failed to fetch hb status:', error);
    }
  };

  const fetchHbConfig = async () => {
    try {
      const res = await api.get('/heartbeat/config');
      if (res.data.code === 0) {
        setHbConfig(res.data.data);
        form.setFieldsValue(res.data.data);
      }
    } catch (error) {
      console.error('Failed to fetch hb config:', error);
    }
  };

  const updateHbConfig = async (values) => {
    setConfigLoading(true);
    try {
      const res = await api.put('/heartbeat/config', values);
      if (res.data.code === 0) {
        setHbConfig(values);
        setIsConfigModalOpen(false);
      }
    } catch (error) {
      console.error('Failed to update hb config:', error);
    } finally {
      setConfigLoading(false);
    }
  };

  const fetchTrendData = async () => {
    try {
      const res = await api.get('/statistics/trend');
      if (res.data.code === 0) {
        setTrendData(res.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch trend:', error);
    }
  };

  const openTrendDrawer = async () => {
    await fetchTrendData();
    setIsTrendDrawerOpen(true);
  };

  const fetchData = async () => {
    try {
      if (selectedStationIDs.length === 0) {
        // No need to fetch if empty, but we can set data to empty to be safe
        setData([]);
        return;
      }
      const params = { limit: limit };
      if (selectedStationIDs.length > 0) {
        params.station_ids = selectedStationIDs.join(',');
      }
      const res = await api.get('/data', { params });
      if (res.data.code === 0) {
        const rawData = res.data.data || [];
        // Ensure values is parsed if it's a string
        const processedData = rawData.map(item => {
          if (typeof item.values === 'string') {
            try {
              item.values = JSON.parse(item.values);
            } catch (e) {
              item.values = [];
            }
          }
          return item;
        });
        setData(processedData);
      }
    } catch (error) {
      console.error('Failed to fetch data:', error);
    }
  };

  const handleDeviceChange = (checkedValues) => {
    setSelectedStationIDs(checkedValues);
  };

  const columns = [
    {
      title: '采集时间',
      dataIndex: 'values',
      key: 'send_time',
      width: 180,
      render: (values) => {
        if (!Array.isArray(values)) return '-';
        const sendTime = values.find(v => v.tag === 'send_time');
        return sendTime && sendTime.value && sendTime.value.string ? sendTime.value.string : '-';
      },
    },
    {
      title: '入库时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (text) => moment(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '设备站码',
      dataIndex: 'device_id',
      key: 'device_id',
      width: 120,
      render: (text) => <Tag color="var(--ant-primary-color)">{text}</Tag>,
    },
    {
      title: '报文内容 & 解析结果',
      key: 'content',
      render: (_, record) => {
        // Prepare Raw Data
        const rawText = record.raw_data || '-';

        // Prepare Parsed Data
        const values = record.values;
        let parsedTags = null;
        if (Array.isArray(values)) {
          const keyMapping = {
            'water_level': '水位',
            'voltage': '电压',
            'cumulative_flow': '累计流量',
            'total_rainfall': '总雨量',
            'rainfall': '雨量',
            'instant_flow': '瞬时流量',
            'hourly_rainfall': '小时雨量',
            'day_rainfall': '日雨量',
            'period_rainfall': '时段雨量',
            'extra_rainfall': '额外雨量',
          };
          parsedTags = (
            <Space size={[4, 4]} wrap style={{ marginTop: 8 }}>
              {values.map((kp, index) => {
                const label = keyMapping[kp.tag] || kp.tag;
                if (kp.tag === 'station_id' || kp.tag === 'send_time') return null;
                if (!kp.value) return null;

                let val = kp.value.float ?? kp.value.int ?? kp.value.string ?? kp.value.bool;
                if (typeof val === 'number' && !Number.isInteger(val)) {
                  val = val.toFixed(3);
                }
                return (
                  <Tag key={index} color="blue" style={{ margin: 0, fontSize: '12px' }}>
                    {label}: {val}
                  </Tag>
                );
              })}
            </Space>
          );
        }

        return (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
            <Text code copyable style={{ maxWidth: '100%', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{rawText}</Text>
            {parsedTags}
          </div>
        );
      },
    },
  ];

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1890ff',
        },
      }}
    >
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{ display: 'flex', alignItems: 'center', background: '#001529', padding: '0 20px' }}>
          <DatabaseOutlined style={{ color: '#fff', fontSize: '24px', marginRight: '10px' }} />
          <Title level={3} style={{ color: '#fff', margin: 0 }}>SL651 监测与解析平台</Title>
        </Header>
        <Layout hasSider>
          <Sider width={280} style={{ background: colorBgContainer, padding: '20px', borderRight: '1px solid #f0f0f0', overflowY: 'auto' }}>
            <div style={{ marginBottom: '20px', display: 'flex', alignItems: 'center' }}>
              <HddOutlined style={{ marginRight: '8px', fontSize: '18px' }} />
              <Title level={4} style={{ margin: 0 }}>设备分组管家</Title>
            </div>
            <Divider style={{ margin: '10px 0' }} orientation="left">在线设备列表</Divider>
            {devices.length > 0 ? (
              <Checkbox.Group
                style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}
                value={selectedStationIDs}
                onChange={handleDeviceChange}
              >
                {devices.map(dev => {
                  const status = hbStatus[dev.id];
                  const isOnline = status ? status.online : (dev.status === 'online');
                  return (
                    <Card
                      key={dev.id}
                      size="small"
                      hoverable
                      style={{
                        borderLeft: `4px solid ${isOnline ? '#52c41a' : '#bfbfbf'}`,
                        marginBottom: '8px'
                      }}
                      bodyStyle={{ padding: '8px' }}
                    >
                      <Checkbox value={dev.id} style={{ width: '100%', marginLeft: 0 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', width: '100%' }}>
                          <Space direction="vertical" size={0}>
                            <Text strong>{dev.name || '未命名设备'}</Text>
                            <Text type="secondary" style={{ fontSize: '11px' }}>ID: {dev.id}</Text>
                          </Space>
                          <Badge
                            status={isOnline ? 'processing' : 'default'}
                            text={<Text style={{ fontSize: '11px' }} type={isOnline ? 'success' : 'secondary'}>{isOnline ? '在线' : '离线'}</Text>}
                          />
                        </div>
                      </Checkbox>
                    </Card>
                  );
                })}
              </Checkbox.Group>
            ) : (
              <Empty description="暂无在线设备" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Sider>
          <Content style={{ padding: '24px', overflowY: 'auto', background: '#f0f2f5', display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <Row gutter={16}>
              <Col span={8}>
                <Card bordered={false} style={{ borderRadius: borderRadiusLG }}>
                  <Statistic
                    title="总设备数"
                    value={stats.total}
                    prefix={<HddOutlined />}
                    valueStyle={{ color: '#001529' }}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card bordered={false} style={{ borderRadius: borderRadiusLG }}>
                  <Statistic
                    title="当前在线"
                    value={stats.online}
                    prefix={<SyncOutlined spin={stats.online > 0} />}
                    valueStyle={{ color: '#52c41a' }}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card bordered={false} style={{ borderRadius: borderRadiusLG }}>
                  <Statistic
                    title="离线警告"
                    value={stats.offline}
                    prefix={<HistoryOutlined />}
                    valueStyle={{ color: stats.offline > 0 ? '#ff4d4f' : '#8c8c8c' }}
                  />
                </Card>
              </Col>
            </Row>

            <Card
              bordered={false}
              style={{ borderRadius: borderRadiusLG, boxShadow: '0 1px 2px 0 rgba(0,0,0,0.03)', flex: 1, display: 'flex', flexDirection: 'column' }}
              bodyStyle={{ padding: '0', flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}
            >
              <div style={{ padding: '16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space>
                  <SyncOutlined spin={loading} />
                  <Title level={5} style={{ margin: 0 }}>实时报文</Title>
                  {selectedStationIDs.length > 0 ? <Tag color="processing">已选: {selectedStationIDs.length}</Tag> : <Tag>全部</Tag>}
                </Space>
                <Space size="large">
                  <Space>
                    <Text secondary style={{ fontSize: '12px' }}>条数:</Text>
                    <input
                      type="range"
                      min="10"
                      max="100"
                      value={limit}
                      onChange={(e) => setLimit(parseInt(e.target.value))}
                      style={{ width: '80px' }}
                    />
                    <Tag style={{ margin: 0 }}>{limit}</Tag>
                  </Space>
                  <Space>
                    <Text secondary style={{ fontSize: '12px' }}>频率:</Text>
                    <select
                      value={refreshInterval}
                      onChange={(e) => setRefreshInterval(parseInt(e.target.value))}
                      style={{ padding: '2px 4px', borderRadius: '4px', border: '1px solid #d9d9d9', fontSize: '12px' }}
                    >
                      <option value={5}>5s</option>
                      <option value={10}>10s</option>
                      <option value={30}>30s</option>
                    </select>
                  </Space>
                  <Tooltip title="查看运行轨迹">
                    <Button
                      type="text"
                      icon={<HistoryOutlined />}
                      onClick={openTrendDrawer}
                    />
                  </Tooltip>
                  <Tooltip title="心跳管理配置">
                    <Button
                      type="text"
                      icon={<SettingOutlined />}
                      onClick={() => setIsConfigModalOpen(true)}
                    />
                  </Tooltip>
                </Space>
              </div>
              <Table
                columns={columns}
                dataSource={data}
                rowKey="id"
                pagination={false}
                size="middle"
                style={{ flex: 1 }}
                scroll={{ y: 'calc(100vh - 360px)' }}
              />
            </Card>

            <Modal
              title={<span><SettingOutlined /> 心跳监测全球配置</span>}
              open={isConfigModalOpen}
              onOk={() => form.submit()}
              onCancel={() => setIsConfigModalOpen(false)}
              confirmLoading={configLoading}
              okText="保存配置"
              cancelText="取消"
            >
              <Form
                form={form}
                layout="vertical"
                onFinish={updateHbConfig}
                initialValues={hbConfig}
              >
                <Form.Item
                  name="check_interval"
                  label="监测巡检间隔 (秒)"
                  extra="系统对所有设备进行活跃度盘点的时间间隔"
                  rules={[{ required: true, message: '请输入巡检间隔' }]}
                >
                  <InputNumber min={5} max={3600} style={{ width: '100%' }} />
                </Form.Item>
                <Form.Item
                  name="default_timeout"
                  label="默认超时判定 (秒)"
                  extra="超过此时间未收到任何报文即判定为离线"
                  rules={[{ required: true, message: '请输入判定超时' }]}
                >
                  <InputNumber min={10} max={86400} style={{ width: '100%' }} />
                </Form.Item>
              </Form>
            </Modal>
            <Drawer
              title={<span><HistoryOutlined /> 设备运行轨迹 (最近24小时)</span>}
              placement="right"
              onClose={() => setIsTrendDrawerOpen(false)}
              open={isTrendDrawerOpen}
              width={400}
            >
              {trendData.length > 0 ? (
                <Timeline
                  items={trendData.map((item, index) => ({
                    color: item.status === 'online' ? 'green' : 'gray',
                    children: (
                      <div>
                        <Text strong>[{item.device_id}]</Text> 变为 <Tag color={item.status === 'online' ? 'success' : 'default'}>{item.status === 'online' ? '在线' : '离线'}</Tag>
                        <br />
                        <Text type="secondary" style={{ fontSize: '12px' }}>{moment(item.time).format('MM-DD HH:mm:ss')}</Text>
                      </div>
                    ),
                  })).reverse()}
                />
              ) : (
                <Empty description="暂无历史轨迹记录" />
              )}
            </Drawer>
          </Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
};

export default App;
