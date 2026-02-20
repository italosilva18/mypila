import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { 
  Wallet, 
  Shield, 
  Zap, 
  BarChart3, 
  Receipt, 
  ArrowRight,
  CheckCircle2,
  Building2,
  Sparkles,
  Star,
  Play
} from 'lucide-react'

export function LandingPage() {
  const navigate = useNavigate()
  const [scrollY, setScrollY] = useState(0)
  const [isVisible, setIsVisible] = useState<{[key: string]: boolean}>({})

  useEffect(() => {
    const handleScroll = () => setScrollY(window.scrollY)
    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          setIsVisible((prev) => ({
            ...prev,
            [entry.target.id]: entry.isIntersecting
          }))
        })
      },
      { threshold: 0.1 }
    )

    document.querySelectorAll('[data-animate]').forEach((el) => {
      observer.observe(el)
    })

    return () => observer.disconnect()
  }, [])

  const features = [
    {
      icon: Wallet,
      title: 'Controle Financeiro Total',
      description: 'Gerencie receitas, despesas e fluxo de caixa em um só lugar. Visualize seu dinheiro de forma clara e objetiva.'
    },
    {
      icon: Building2,
      title: 'Gestão de Empresas',
      description: 'Cadastre múltiplas empresas e tenha controle individualizado de cada uma. Ideal para empreendedores com diversos negócios.'
    },
    {
      icon: BarChart3,
      title: 'Relatórios Inteligentes',
      description: 'Gráficos e relatórios automáticos que mostram exatamente onde seu dinheiro está indo e como está o desempenho.'
    },
    {
      icon: Receipt,
      title: 'Controle de Transações',
      description: 'Registre todas as movimentações financeiras com categorização automática e lembretes de pagamento.'
    },
    {
      icon: Shield,
      title: 'Segurança de Dados',
      description: 'Seus dados protegidos com criptografia de ponta. Acesso seguro com autenticação JWT e controle de permissões.'
    },
    {
      icon: Zap,
      title: 'Automatização',
      description: 'Automatize tarefas repetitivas, configure lembretes de pagamento e receba notificações importantes.'
    }
  ]

  const benefits = [
    'Economize até 40% nas despesas mensais',
    'Reduza o tempo de gestão financeira em 80%',
    'Tome decisões baseadas em dados reais',
    'Evite esquecimento de contas a pagar',
    'Tenha previsibilidade financeira',
    'Acesso de qualquer lugar, a qualquer hora'
  ]

  const stats = [
    { value: 'R$ 10M+', label: 'Em transações gerenciadas' },
    { value: '500+', label: 'Empresas cadastradas' },
    { value: '2.000+', label: 'Usuários ativos' },
    { value: '99.9%', label: 'Uptime do sistema' }
  ]

  const howItWorks = [
    {
      step: '1',
      title: 'Crie sua conta',
      description: 'Cadastre-se gratuitamente em menos de 2 minutos. Não precisa de cartão de crédito.'
    },
    {
      step: '2',
      title: 'Cadastre sua empresa',
      description: 'Adicione suas empresas e configure as informações básicas. Você pode ter quantas empresas quiser.'
    },
    {
      step: '3',
      title: 'Comece a controlar',
      description: 'Registre receitas e despesas, visualize relatórios e tenha total controle financeiro.'
    }
  ]

  return (
    <div className="min-h-screen bg-white overflow-x-hidden">
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-white/80 backdrop-blur-lg border-b border-gray-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-2">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-indigo-600 rounded-xl flex items-center justify-center">
                <Wallet className="text-white" size={24} />
              </div>
              <span className="text-xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                MyPila
              </span>
            </div>
            <div className="flex items-center gap-4">
              <button 
                onClick={() => navigate('/login')}
                className="text-gray-600 hover:text-gray-900 font-medium transition-colors"
              >
                Entrar
              </button>
              <button 
                onClick={() => navigate('/login')}
                className="bg-gradient-to-r from-blue-600 to-indigo-600 text-white px-6 py-2.5 rounded-lg font-semibold hover:shadow-lg hover:shadow-blue-500/30 transition-all transform hover:-translate-y-0.5"
              >
                Começar Grátis
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 lg:pt-48 lg:pb-32 overflow-hidden">
        {/* Background Effects */}
        <div className="absolute inset-0 overflow-hidden">
          <div 
            className="absolute -top-40 -right-40 w-96 h-96 bg-blue-400/20 rounded-full blur-3xl"
            style={{ transform: `translateY(${scrollY * 0.3}px)` }}
          />
          <div 
            className="absolute top-1/2 -left-40 w-96 h-96 bg-indigo-400/20 rounded-full blur-3xl"
            style={{ transform: `translateY(${scrollY * -0.2}px)` }}
          />
          <div 
            className="absolute bottom-0 right-1/4 w-64 h-64 bg-purple-400/20 rounded-full blur-3xl"
            style={{ transform: `translateY(${scrollY * 0.1}px)` }}
          />
        </div>

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            {/* Badge */}
            <div 
              data-animate
              id="hero-badge"
              className={`inline-flex items-center gap-2 px-4 py-2 rounded-full bg-blue-50 border border-blue-100 mb-8 transition-all duration-700 ${
                isVisible['hero-badge'] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
              }`}
            >
              <Sparkles className="text-blue-600" size={16} />
              <span className="text-sm font-medium text-blue-700">Sistema de Gestão Financeira #1 do Brasil</span>
            </div>

            {/* Main Headline */}
            <h1 
              data-animate
              id="hero-title"
              className={`text-5xl sm:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 transition-all duration-700 delay-100 ${
                isVisible['hero-title'] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
              }`}
            >
              <span className="text-gray-900">Domine suas</span>
              <br />
              <span className="bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 bg-clip-text text-transparent">
                Finanças Empresariais
              </span>
            </h1>

            {/* Subtitle */}
            <p 
              data-animate
              id="hero-subtitle"
              className={`text-xl text-gray-600 max-w-2xl mx-auto mb-10 transition-all duration-700 delay-200 ${
                isVisible['hero-subtitle'] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
              }`}
            >
              O MyPila é o sistema completo de gestão financeira que você precisa para 
              <span className="font-semibold text-gray-900"> controlar receitas, despesas e o fluxo de caixa </span>
              da sua empresa de forma simples e eficiente.
            </p>

            {/* CTA Buttons */}
            <div 
              data-animate
              id="hero-cta"
              className={`flex flex-col sm:flex-row items-center justify-center gap-4 mb-16 transition-all duration-700 delay-300 ${
                isVisible['hero-cta'] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
              }`}
            >
              <button 
                onClick={() => navigate('/login')}
                className="group w-full sm:w-auto bg-gradient-to-r from-blue-600 to-indigo-600 text-white px-8 py-4 rounded-xl font-bold text-lg hover:shadow-2xl hover:shadow-blue-500/40 transition-all transform hover:-translate-y-1 flex items-center justify-center gap-2"
              >
                <Play size={20} className="fill-current" />
                Começar Agora - É Grátis!
                <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
              </button>
              <button className="w-full sm:w-auto px-8 py-4 rounded-xl font-semibold text-gray-700 border-2 border-gray-200 hover:border-gray-300 hover:bg-gray-50 transition-all flex items-center justify-center gap-2">
                <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
                  <Play size={16} className="text-gray-600 fill-current ml-0.5" />
                </div>
                Ver Demonstração
              </button>
            </div>

            {/* Trust Badges */}
            <div 
              data-animate
              id="hero-trust"
              className={`flex flex-wrap items-center justify-center gap-6 text-sm text-gray-500 transition-all duration-700 delay-400 ${
                isVisible['hero-trust'] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
              }`}
            >
              <div className="flex items-center gap-2">
                <CheckCircle2 className="text-green-500" size={18} />
                <span>Gratuito para começar</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="text-green-500" size={18} />
                <span>Sem cartão de crédito</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="text-green-500" size={18} />
                <span>Cancelamento fácil</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-12 bg-gradient-to-r from-gray-900 to-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-8">
            {stats.map((stat, index) => (
              <div 
                key={index} 
                data-animate
                id={`stat-${index}`}
                className={`text-center transition-all duration-700 ${
                  isVisible[`stat-${index}`] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
                }`}
                style={{ transitionDelay: `${index * 100}ms` }}
              >
                <div className="text-3xl lg:text-4xl font-bold text-white mb-2">{stat.value}</div>
                <div className="text-gray-400 text-sm">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-24 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <span className="text-blue-600 font-semibold text-sm uppercase tracking-wider">Recursos Poderosos</span>
            <h2 className="text-4xl font-bold text-gray-900 mt-3 mb-4">
              Tudo que você precisa para<br />
              <span className="text-blue-600">gerenciar suas finanças</span>
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Ferramentas completas e intuitivas para você ter total controle do dinheiro da sua empresa
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <div 
                key={index}
                data-animate
                id={`feature-${index}`}
                className={`group bg-white rounded-2xl p-8 shadow-sm hover:shadow-xl transition-all duration-500 border border-gray-100 hover:border-blue-200 ${
                  isVisible[`feature-${index}`] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
                }`}
                style={{ transitionDelay: `${index * 100}ms` }}
              >
                <div className="w-14 h-14 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-300 shadow-lg shadow-blue-500/30">
                  <feature.icon className="text-white" size={28} />
                </div>
                <h3 className="text-xl font-bold text-gray-900 mb-3">{feature.title}</h3>
                <p className="text-gray-600 leading-relaxed">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-24 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <span className="text-blue-600 font-semibold text-sm uppercase tracking-wider">Simples e Rápido</span>
            <h2 className="text-4xl font-bold text-gray-900 mt-3 mb-4">
              Comece em <span className="text-blue-600">3 passos</span>
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Configure seu sistema financeiro em menos de 5 minutos
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {howItWorks.map((step, index) => (
              <div 
                key={index}
                data-animate
                id={`step-${index}`}
                className={`relative text-center transition-all duration-700 ${
                  isVisible[`step-${index}`] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
                }`}
                style={{ transitionDelay: `${index * 150}ms` }}
              >
                {index < 2 && (
                  <div className="hidden md:block absolute top-12 left-1/2 w-full h-0.5 bg-gradient-to-r from-blue-200 to-transparent" />
                )}
                <div className="w-24 h-24 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white text-3xl font-bold shadow-xl shadow-blue-500/30">
                  {step.step}
                </div>
                <h3 className="text-xl font-bold text-gray-900 mb-3">{step.title}</h3>
                <p className="text-gray-600">{step.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Benefits Section */}
      <section className="py-24 bg-gradient-to-br from-blue-50 to-indigo-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid lg:grid-cols-2 gap-16 items-center">
            <div>
              <span className="text-blue-600 font-semibold text-sm uppercase tracking-wider">Benefícios</span>
              <h2 className="text-4xl font-bold text-gray-900 mt-3 mb-6">
                Por que empresários escolhem o MyPila?
              </h2>
              <p className="text-xl text-gray-600 mb-8">
                Milhares de empresários já descobriram como é fácil manter as finanças organizadas 
                e tomar decisões melhores para seus negócios.
              </p>
              <button 
                onClick={() => navigate('/login')}
                className="group bg-gradient-to-r from-blue-600 to-indigo-600 text-white px-8 py-4 rounded-xl font-bold text-lg hover:shadow-2xl hover:shadow-blue-500/40 transition-all transform hover:-translate-y-1 flex items-center gap-2"
              >
                Começar Agora
                <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
              </button>
            </div>

            <div className="grid gap-4">
              {benefits.map((benefit, index) => (
                <div 
                  key={index}
                  data-animate
                  id={`benefit-${index}`}
                  className={`flex items-center gap-4 bg-white p-4 rounded-xl shadow-sm transition-all duration-500 ${
                    isVisible[`benefit-${index}`] ? 'opacity-100 translate-x-0' : 'opacity-0 -translate-x-8'
                  }`}
                  style={{ transitionDelay: `${index * 100}ms` }}
                >
                  <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center flex-shrink-0">
                    <CheckCircle2 className="text-green-600" size={20} />
                  </div>
                  <span className="font-medium text-gray-800">{benefit}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section className="py-24 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <span className="text-blue-600 font-semibold text-sm uppercase tracking-wider">Depoimentos</span>
            <h2 className="text-4xl font-bold text-gray-900 mt-3 mb-4">
              O que nossos usuários dizem
            </h2>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {[
              {
                name: 'Roberto Silva',
                role: 'Empresário',
                text: 'O MyPila mudou completamente como gerencio as finanças da minha empresa. Agora sei exatamente para onde vai cada real.',
                rating: 5
              },
              {
                name: 'Ana Paula',
                role: 'Contadora',
                text: 'Recomendo para todos meus clientes. A facilidade de uso e os relatórios detalhados fazem toda a diferença.',
                rating: 5
              },
              {
                name: 'Carlos Mendes',
                role: 'Startup Founder',
                text: 'Finalmente encontrei uma ferramenta que entende as necessidades de quem tem múltiplos negócios. Sensacional!',
                rating: 5
              }
            ].map((testimonial, index) => (
              <div 
                key={index}
                data-animate
                id={`testimonial-${index}`}
                className={`bg-gray-50 rounded-2xl p-8 transition-all duration-700 ${
                  isVisible[`testimonial-${index}`] ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
                }`}
                style={{ transitionDelay: `${index * 150}ms` }}
              >
                <div className="flex gap-1 mb-4">
                  {[...Array(testimonial.rating)].map((_, i) => (
                    <Star key={i} className="text-yellow-400 fill-current" size={20} />
                  ))}
                </div>
                <p className="text-gray-700 mb-6 italic">"{testimonial.text}"</p>
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold">
                    {testimonial.name.charAt(0)}
                  </div>
                  <div>
                    <div className="font-semibold text-gray-900">{testimonial.name}</div>
                    <div className="text-sm text-gray-500">{testimonial.role}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-24 bg-gradient-to-br from-blue-600 via-indigo-600 to-purple-600 relative overflow-hidden">
        {/* Background Pattern */}
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-0 left-0 w-96 h-96 bg-white rounded-full blur-3xl" />
          <div className="absolute bottom-0 right-0 w-96 h-96 bg-white rounded-full blur-3xl" />
        </div>

        <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-4xl lg:text-5xl font-bold text-white mb-6">
            Pronto para organizar suas finanças?
          </h2>
          <p className="text-xl text-blue-100 mb-10 max-w-2xl mx-auto">
            Junte-se a milhares de empresários que já descobriram uma forma 
            mais simples de gerenciar o dinheiro do seu negócio.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <button 
              onClick={() => navigate('/login')}
              className="group w-full sm:w-auto bg-white text-blue-600 px-10 py-5 rounded-xl font-bold text-lg hover:shadow-2xl transition-all transform hover:-translate-y-1 flex items-center justify-center gap-2"
            >
              Criar Conta Grátis
              <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
            </button>
            <button className="w-full sm:w-auto px-8 py-5 rounded-xl font-semibold text-white border-2 border-white/30 hover:bg-white/10 transition-all">
              Falar com Suporte
            </button>
          </div>
          <p className="text-blue-200 mt-6 text-sm">
            ✓ Gratuito para começar ✓ Sem cartão de crédito ✓ Cancele quando quiser
          </p>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-300 py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid md:grid-cols-4 gap-8 mb-12">
            <div className="col-span-2">
              <div className="flex items-center gap-2 mb-4">
                <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl flex items-center justify-center">
                  <Wallet className="text-white" size={24} />
                </div>
                <span className="text-xl font-bold text-white">MyPila</span>
              </div>
              <p className="text-gray-400 mb-4 max-w-sm">
                Sistema completo de gestão financeira empresarial. 
                Controle receitas, despesas e tenha total domínio do seu dinheiro.
              </p>
            </div>
            <div>
              <h4 className="text-white font-semibold mb-4">Produto</h4>
              <ul className="space-y-2">
                <li><a href="#" className="hover:text-white transition-colors">Funcionalidades</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Preços</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Integrações</a></li>
                <li><a href="#" className="hover:text-white transition-colors">API</a></li>
              </ul>
            </div>
            <div>
              <h4 className="text-white font-semibold mb-4">Suporte</h4>
              <ul className="space-y-2">
                <li><a href="#" className="hover:text-white transition-colors">Central de Ajuda</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Documentação</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Contato</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Status</a></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-gray-800 pt-8 flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-gray-500 text-sm">
              © 2026 MyPila. Todos os direitos reservados.
            </p>
            <div className="flex items-center gap-6 text-sm">
              <a href="#" className="hover:text-white transition-colors">Termos de Uso</a>
              <a href="#" className="hover:text-white transition-colors">Privacidade</a>
              <a href="#" className="hover:text-white transition-colors">Cookies</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
