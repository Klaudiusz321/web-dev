import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, Globe, Sparkles, Target, Search, Shield } from 'lucide-react'
import { useCreateUrl } from '../hooks/useUrls'
import toast from 'react-hot-toast'

export default function AddUrl() {
  const navigate = useNavigate()
  const [url, setUrl] = useState('')
  const createUrlMutation = useCreateUrl()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!url.trim()) return

    try {
      // Validate URL format
      new URL(url)
      
      // Create URL and start crawling
      await createUrlMutation.mutateAsync(url)
      
      toast.success('URL added successfully! Crawling started.')
      
      // Navigate back to dashboard
      navigate('/')
    } catch (error) {
      if (error instanceof TypeError) {
        toast.error('Please enter a valid URL format')
      } else {
        toast.error('Failed to add URL. Please try again.')
      }
    }
  }

  const isValidUrl = (urlString: string) => {
    try {
      new URL(urlString)
      return true
    } catch {
      return false
    }
  }

  const analysisFeatures = [
    {
      icon: Target,
      title: 'SEO Analysis',
      description: 'HTML version, page title, and meta tags'
    },
    {
      icon: Search,
      title: 'Content Structure',
      description: 'Heading hierarchy (H1-H6) and content organization'
    },
    {
      icon: Globe,
      title: 'Link Discovery',
      description: 'Internal vs external links and link validation'
    },
    {
      icon: Shield,
      title: 'Accessibility Check',
      description: 'Broken links detection and accessibility issues'
    }
  ]

  return (
    <div className="flex-1 p-8 bg-bg-primary min-h-screen">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate('/')}
            className="flex items-center text-text-tertiary hover:text-text-primary transition-colors group mb-6"
          >
            <ArrowLeft className="w-4 h-4 mr-2 group-hover:-translate-x-0.5 transition-transform" />
            Back to Dashboard
          </button>

          <div className="flex items-center space-x-4 mb-4">
            <div className="w-12 h-12 bg-gradient-to-br from-accent to-accent-secondary rounded-2xl flex items-center justify-center">
              <Plus className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-text-primary">Add New URL</h1>
              <p className="text-text-secondary mt-1">
                Start crawling and analyzing any website with our AI-powered web crawler
              </p>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Form */}
          <div className="lg:col-span-2">
            <div className="bg-bg-secondary/50 backdrop-blur-sm border border-border rounded-2xl p-8 card-hover">
              <div className="flex items-center space-x-3 mb-6">
                <Globe className="w-6 h-6 text-accent" />
                <h2 className="text-xl font-semibold text-text-primary">Website URL</h2>
              </div>

              <form onSubmit={handleSubmit} className="space-y-6">
                <div>
                  <label htmlFor="url" className="block text-sm font-medium text-text-primary mb-3">
                    Enter the website URL you want to analyze
                  </label>
                  <div className="relative">
                    <input
                      type="url"
                      id="url"
                      value={url}
                      onChange={(e) => setUrl(e.target.value)}
                      placeholder="https://your-amazing-website.com"
                      className="w-full px-4 py-4 bg-bg-primary border border-border rounded-xl focus:ring-2 focus:ring-accent/50 focus:border-accent transition-all text-text-primary placeholder:text-text-tertiary text-lg"
                      required
                    />
                    {url && !isValidUrl(url) && (
                      <p className="mt-3 text-sm text-error flex items-center">
                        <span className="w-2 h-2 bg-error rounded-full mr-2"></span>
                        Please enter a valid URL starting with http:// or https://
                      </p>
                    )}
                  </div>
                </div>

                <div className="flex flex-col sm:flex-row gap-3">
                  <button
                    type="button"
                    onClick={() => navigate('/')}
                    className="px-6 py-3 text-text-secondary bg-bg-tertiary border border-border rounded-xl hover:bg-bg-tertiary/70 hover:text-text-primary transition-all"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={!url.trim() || !isValidUrl(url) || createUrlMutation.isPending}
                    className="flex-1 flex items-center justify-center px-6 py-3 bg-gradient-to-r from-accent to-accent-secondary text-white rounded-xl hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg"
                  >
                    {createUrlMutation.isPending ? (
                      <>
                        <div className="animate-spin rounded-full h-5 w-5 border-2 border-white border-t-transparent mr-3"></div>
                        Starting Analysis...
                      </>
                    ) : (
                      <>
                        <Sparkles className="w-5 h-5 mr-3" />
                        Start Web Analysis
                      </>
                    )}
                  </button>
                </div>
              </form>
            </div>
          </div>

          {/* Analysis Features */}
          <div className="space-y-6">
            <div className="bg-bg-secondary/50 backdrop-blur-sm border border-border rounded-2xl p-6">
              <h3 className="text-lg font-semibold text-text-primary mb-4 flex items-center">
                <Sparkles className="w-5 h-5 text-accent mr-2" />
                What We'll Analyze
              </h3>
              
              <div className="space-y-4">
                {analysisFeatures.map((feature, index) => (
                  <div key={index} className="flex items-start space-x-3 p-3 rounded-xl hover:bg-bg-tertiary/30 transition-colors">
                    <div className="w-8 h-8 bg-accent/10 rounded-lg flex items-center justify-center flex-shrink-0">
                      <feature.icon className="w-4 h-4 text-accent" />
                    </div>
                    <div>
                      <h4 className="font-medium text-text-primary text-sm">{feature.title}</h4>
                      <p className="text-text-tertiary text-xs mt-1">{feature.description}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

          </div>
        </div>
      </div>
    </div>
  )
} 