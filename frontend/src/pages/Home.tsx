import {createLazyRoute} from "@tanstack/react-router";
import {
  Box,
  Title,
  Text,
  Button,
  Container,
  SimpleGrid,
  Card,
  ThemeIcon,
  Space,
  Divider,
} from "@mantine/core";
import {
  AiOutlineFund,
  AiOutlineInfo,
  AiOutlineLineChart,
  AiOutlineLock,
  AiOutlineRocket,
  AiOutlineWallet
} from "react-icons/ai";
import {IoSchoolOutline} from "react-icons/io5";
import {Header} from "@local/components/Header.tsx";
import {Footer} from "@local/components/Footer.tsx";

function Home() {
  return (
    <>
      <Box style={{
        display: 'flex',
        flexDirection: 'column',
        minHeight: '100vh'
      }}>
      <Header/>
      <Box component={'main'} flex={1}>
        <Box style={{ textAlign: 'center', padding: '40px 1rem' }}>
          <Title order={1} style={{ fontSize: '2.5rem', marginBottom: '1rem' }}>
            Practice Trading, Master the Markets, Risk-Free.
          </Title>
          <Text size={"lg"} c={'dimmed'} style={{marginBottom: '2rem'}}>
            Hone your trading skills with real-time market data and virtual funds. Top up your practice capital easily and affordably.
          </Text>
          <Button size={"lg"}>Start Trading Now</Button>
        </Box>

        <Container py={{ lg: 'md', md: 'lg', sm: 'xl' }} id={'features'} style={{ paddingLeft: 0, paddingRight: 0 }}>
          <Box>
            <Title order={2} mb={'xl'} style={{textAlign: 'center'}}>
              Why Choose PaperTrading?
            </Title>
            <SimpleGrid cols={{xl: 4, md: 3, sm: 2, xs: 1}} spacing={'xl'}>
              {[
                { title: "Real Market Data", desc: "Trade with live data from global markets, just like the pros.", icon: <AiOutlineLineChart/> },
                { title: "Virtual Funds & Easy Top-Ups", desc: "Start with a generous virtual balance and replenish affordably if needed.", icon: <AiOutlineWallet/> },
                { title: "Realistic Simulation", desc: "Experience true market conditions and test your strategies effectively.", icon: <AiOutlineRocket/> },
                { title: "Educational Resources", desc: "Access guides and tutorials to boost your trading knowledge.", icon: <IoSchoolOutline/> },
                { title: "Portfolio Tracking", desc: "Monitor your virtual investments and track your performance.", icon: <AiOutlineFund/> },
                { title: "Risk-Free Practice", desc: "Learn and experiment without any financial risk.", icon: <AiOutlineLock/> },
                { title: "Direct Market Info", desc: "Our platform uses information taken directly from the real market(s).", icon: <AiOutlineInfo/> }
              ].map(feature => (
                <Card shadow={'sm'} p={'lg'} key={feature.title}>
                  <ThemeIcon size="lg" variant="light">{feature.icon}</ThemeIcon>
                  <Space h={'xs'}></Space>
                  <Text size={'lg'} style={{fontWeight: "bold"}}>{feature.title}</Text>
                  <Text size={'sm'} c={'dimmed'}>{feature.desc}</Text>
                </Card>
              ))}
            </SimpleGrid>
          </Box>
        </Container>

        <Divider/>

        <Box id="how-it-works" py={'xl'}>
          <Container>
            <Title style={{textAlign: 'center'}} order={2} mb={'xl'}>
              Get Started in 3 Simple Steps
            </Title>
            <SimpleGrid cols={{md: 3, sm: 1}} spacing="lg">
              {[
                { num: "1", title: "Sign Up", desc: "Create your free PaperTrading account in minutes." },
                { num: "2", title: "Practice", desc: "Access real-time markets, use your virtual funds, and place trades." },
                { num: "3", title: "Learn & Grow", desc: "Utilize our resources, track your progress, and refine your strategies." }
              ].map(step => (
                <Card p={'xl'} shadow={'sm'} key={step.num}>
                  <Text size="xl" c="blue" style={{fontWeight: 700}}>{step.num}.</Text>
                  <Text mt="sm" mb="xs" size="lg" style={{fontWeight: 600}}>{step.title}</Text>
                  <Text size="sm" c="dimmed">{step.desc}</Text>
                </Card>
              ))}
            </SimpleGrid>
          </Container>
        </Box>

        <Divider/>

        <Container py="xl" style={{ textAlign: 'center' }}>
          <Title order={2} mb={'mb'}>Ready to Start Your Trading Journey?</Title>
          <Text c={'dimmed'} mb={"xl"} style={{maxWidth: '600px', margin: '0 auto 1.5rem auto'}}>
            Join thousands of aspiring traders who are building their skills with PaperTrading. It's free to get started!
          </Text>
          <Button size="xl" className="cta-button final-cta">Sign Up for Free</Button>
        </Container>
      </Box>
      <Footer/>
      </Box>
    </>
  )
}

export const Route = createLazyRoute("/")({
  component: Home,
})
