import React from 'react';

export default function Home() {
  return (
    <div dangerouslySetInnerHTML={{ __html: `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>COO-LLM - Chief Operations Officer for Your LLMs</title>

    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <!-- Font Awesome -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <!-- Google Fonts -->
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700;900&family=Orbitron:wght@400;700;900&family=Rajdhani:wght@500;700&display=swap" rel="stylesheet">

    <style>
        :root {
            --primary-red: #DC143C;
            --navy-blue: #0F172A;
            --light-blue: #1E293B;
            --accent-gold: #FFD700;
            --dark-text: #0D1117;
            --light-text: #E6E6E6;
            --card-bg: rgba(30, 41, 59, 0.7);
        }

        body {
            font-family: 'Poppins', sans-serif;
            color: var(--light-text);
            overflow-x: hidden;
            background-color: var(--navy-blue);
        }

        h1, h2, h3, h4, h5, h6 {
            font-family: 'Orbitron', sans-serif;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .tech-font {
            font-family: 'Rajdhani', sans-serif;
            font-weight: 700;
        }

        /* Animations */
        @keyframes float {
            0%, 100% { transform: translateY(0); }
            50% { transform: translateY(-10px); }
        }

        @keyframes pulse {
            0%, 100% { transform: scale(1); box-shadow: 0 0 0 0 rgba(220, 20, 60, 0.7); }
            50% { transform: scale(1.05); box-shadow: 0 0 0 10px rgba(220, 20, 60, 0); }
        }

        @keyframes glow {
            0%, 100% { box-shadow: 0 0 5px var(--primary-red), 0 0 10px var(--primary-red); }
            50% { box-shadow: 0 0 20px var(--primary-red), 0 0 30px var(--primary-red); }
        }

        @keyframes typewriter {
            from { width: 0; }
            to { width: 100%; }
        }

        @keyframes blink {
            0%, 100% { opacity: 1; }
            50% { opacity: 0; }
        }

        @keyframes slideInLeft {
            from { transform: translateX(-100px); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }

        @keyframes slideInRight {
            from { transform: translateX(100px); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }

        @keyframes slideInUp {
            from { transform: translateY(50px); opacity: 0; }
            to { transform: translateY(0); opacity: 1; }
        }

        @keyframes zoomIn {
            from { transform: scale(0.8); opacity: 0; }
            to { transform: scale(1); opacity: 1; }
        }

        /* Particle Background */
        #particles-js {
            position: fixed;
            width: 100%;
            height: 100%;
            top: 0;
            left: 0;
            z-index: -1;
        }

        /* Glitch Effect */
        .glitch {
            position: relative;
            color: var(--primary-red);
            font-size: 4rem;
            font-weight: 900;
            text-transform: uppercase;
            text-shadow: 0.05em 0 0 rgba(255, 0, 0, 0.75), -0.025em -0.05em 0 rgba(0, 255, 0, 0.75), 0.025em 0.05em 0 rgba(0, 0, 255, 0.75);
            animation: glitch 500ms infinite;
        }

        .glitch::before, .glitch::after {
            content: attr(data-text);
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
        }

        .glitch::before {
            left: 2px;
            text-shadow: -1px 0 var(--primary-red);
            clip: rect(44px, 450px, 56px, 0);
            animation: glitch-anim 5s infinite linear alternate-reverse;
        }

        .glitch::after {
            left: -2px;
            text-shadow: -1px 0 var(--accent-gold);
            clip: rect(44px, 450px, 56px, 0);
            animation: glitch-anim2 1s infinite linear alternate-reverse;
        }

        @keyframes glitch-anim {
            0% { clip: rect(65px, 9999px, 119px, 0); }
            20% { clip: rect(30px, 9999px, 25px, 0); }
            40% { clip: rect(87px, 9999px, 82px, 0); }
            60% { clip: rect(15px, 9999px, 94px, 0); }
            80% { clip: rect(92px, 9999px, 98px, 0); }
            100% { clip: rect(8px, 9999px, 82px, 0); }
        }

        @keyframes glitch-anim2 {
            0% { clip: rect(32px, 9999px, 85px, 0); }
            20% { clip: rect(54px, 9999px, 73px, 0); }
            40% { clip: rect(12px, 9999px, 23px, 0); }
            60% { clip: rect(63px, 9999px, 27px, 0); }
            80% { clip: rect(34px, 9999px, 55px, 0); }
            100% { clip: rect(81px, 9999px, 73px, 0); }
        }

        /* Neon Text */
        .neon-text {
            color: var(--light-text);
            text-shadow: 0 0 5px var(--primary-red), 0 0 10px var(--primary-red), 0 0 20px var(--primary-red), 0 0 40px var(--primary-red);
        }

        /* Navbar */
        .navbar {
            background-color: rgba(15, 23, 42, 0.9);
            backdrop-filter: blur(10px);
            box-shadow: 0 4px 20px rgba(220, 20, 60, 0.3);
            padding: 15px 0;
            border-bottom: 3px solid var(--primary-red);
            z-index: 1000;
            animation: slideInDown 0.8s ease-out;
        }

        .navbar-brand {
            font-weight: 900;
            font-size: 1.8rem;
            color: var(--primary-red) !important;
            display: flex;
            align-items: center;
            gap: 10px;
            text-shadow: 0 0 10px var(--primary-red);
        }

        .brand-logo {
            width: 40px;
            height: 40px;
            animation: pulse 2s infinite;
        }

        .nav-link {
            font-weight: 700;
            margin: 0 10px;
            color: var(--light-text) !important;
            transition: all 0.3s;
            border-radius: 20px;
            padding: 8px 20px !important;
            position: relative;
            overflow: hidden;
            z-index: 1;
        }

        .nav-link::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 0;
            height: 100%;
            background-color: var(--primary-red);
            z-index: -1;
            transition: width 0.3s ease;
        }

        .nav-link:hover::before {
            width: 100%;
        }

        .nav-link:hover {
            color: var(--navy-blue) !important;
            transform: translateY(-3px);
            box-shadow: 0 5px 15px rgba(220, 20, 60, 0.4);
        }

        /* Buttons */
        .btn-primary {
            background-color: var(--primary-red);
            border-color: var(--primary-red);
            font-weight: 700;
            padding: 12px 30px;
            border-radius: 50px;
            transition: all 0.3s;
            box-shadow: 0 4px 15px rgba(220, 20, 60, 0.4);
            position: relative;
            overflow: hidden;
            z-index: 1;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .btn-primary::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 0;
            height: 100%;
            background-color: var(--navy-blue);
            z-index: -1;
            transition: width 0.3s ease;
        }

        .btn-primary:hover::before {
            width: 100%;
        }

        .btn-primary:hover {
            border-color: var(--navy-blue);
            transform: translateY(-5px);
            box-shadow: 0 10px 25px rgba(220, 20, 60, 0.6);
        }

        .btn-outline-primary {
            color: var(--primary-red);
            border-color: var(--primary-red);
            font-weight: 700;
            padding: 12px 30px;
            border-radius: 50px;
            transition: all 0.3s;
            text-transform: uppercase;
            letter-spacing: 1px;
            position: relative;
            overflow: hidden;
            z-index: 1;
        }

        .btn-outline-primary::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 0;
            height: 100%;
            background-color: var(--primary-red);
            z-index: -1;
            transition: width 0.3s ease;
        }

        .btn-outline-primary:hover::before {
            width: 100%;
        }

        .btn-outline-primary:hover {
            color: var(--navy-blue);
            border-color: var(--primary-red);
            transform: translateY(-5px);
            box-shadow: 0 10px 25px rgba(220, 20, 60, 0.6);
        }

        /* Hero Section */
        .hero-section {
            padding: 120px 0;
            position: relative;
            overflow: hidden;
            background: linear-gradient(135deg, var(--navy-blue) 0%, var(--light-blue) 100%);
        }

        .hero-grid {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-image: linear-gradient(rgba(220, 20, 60, 0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(220, 20, 60, 0.1) 1px, transparent 1px);
            background-size: 50px 50px;
            z-index: -1;
            animation: grid-move 20s linear infinite;
        }

        @keyframes grid-move {
            0% { background-position: 0 0; }
            100% { background-position: 50px 50px; }
        }

        .hero-title {
            font-size: 4rem;
            font-weight: 900;
            margin-bottom: 30px;
            color: var(--primary-red);
            text-shadow: 0 0 15px rgba(220, 20, 60, 0.5);
            animation: slideInLeft 1s ease-out;
            position: relative;
            display: block;
            overflow: visible;
        }

        .hero-title::after {
            content: '';
            position: absolute;
            bottom: -10px;
            left: 0;
            width: 100px;
            height: 5px;
            background: linear-gradient(90deg, var(--primary-red), var(--accent-gold));
            border-radius: 5px;
            animation: glow 2s infinite alternate;
        }

        .typewriter-container {
            overflow: hidden;
            white-space: nowrap;
            margin: 0 auto;
            letter-spacing: .15em;
            animation: typewriter 3.5s steps(80, end), blink .75s step-end infinite;
            display: inline-block;
            width: 100%;
        }

        .hero-subtitle {
            font-size: 1.4rem;
            margin-bottom: 30px;
            color: var(--light-text);
            font-weight: 500;
            animation: slideInLeft 1.2s ease-out;
            display: block;
            overflow: visible;
        }

        .coo-character {
            position: relative;
            text-align: center;
            animation: slideInRight 1s ease-out;
        }

        .coo-image {
            max-width: 90%;
            border-radius: 20px;
            transition: all 0.3s;
        }

        .coo-image:hover {
            transform: scale(1.05);
        }

        .speech-bubble {
            position: absolute;
            top: 60px;
            right: 12px;
            background-color: rgba(15, 23, 42, 0.9);
            border-radius: 20px;
            padding: 15px;
            box-shadow: 0 0 20px rgba(220, 20, 60, 0.5);
            max-width: 200px;
            border: 2px solid var(--primary-red);
            animation: pulse 3s infinite;
        }

        .speech-bubble::after {
            content: '';
            position: absolute;
            bottom: 20px;
            left: -15px;
            width: 30px;
            height: 30px;
            background-color: rgba(15, 23, 42, 0.9);
            border-left: 2px solid var(--primary-red);
            border-bottom: 2px solid var(--primary-red);
            transform: rotate(45deg);
            z-index: -1;
        }

        .speech-text {
            font-family: 'Rajdhani', sans-serif;
            font-size: 1.2rem;
            color: var(--light-text);
            text-transform: uppercase;
            font-weight: 700;
        }

        /* Comic Style Section */
        .comic-section {
            padding: 80px 0;
            position: relative;
            background: linear-gradient(180deg, var(--light-blue) 0%, var(--navy-blue) 100%);
        }

        .section-title {
            font-size: 3rem;
            font-weight: 900;
            margin-bottom: 50px;
            position: relative;
            display: inline-block;
            color: var(--primary-red);
            text-align: center;
            width: 100%;
            text-shadow: 0 0 15px rgba(220, 20, 60, 0.5);
            animation: zoomIn 1s ease-out;
        }

        .section-title::after {
            content: '';
            position: absolute;
            bottom: -15px;
            left: 50%;
            transform: translateX(-50%);
            width: 150px;
            height: 5px;
            background: linear-gradient(90deg, transparent, var(--primary-red), transparent);
            border-radius: 5px;
            animation: glow 2s infinite alternate;
        }

        .comic-panel {
            background-color: var(--card-bg);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            margin-bottom: 30px;
            position: relative;
            border: 3px solid var(--primary-red);
            transition: all 0.5s;
            overflow: visible;
            z-index: 1;
            display: block;
        }

        .comic-panel::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: radial-gradient(circle at center, transparent 0%, rgba(220, 20, 60, 0.1) 100%);
            z-index: -1;
            border-radius: 12px;
        }

        .comic-panel:hover {
            transform: translateY(-15px) rotate(1deg);
            box-shadow: 0 20px 40px rgba(220, 20, 60, 0.4);
            border-color: var(--accent-gold);
        }

        .panel-number {
            position: absolute;
            top: -20px;
            left: 20px;
            background-color: var(--primary-red);
            color: var(--light-text);
            width: 50px;
            height: 50px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1.8rem;
            font-weight: 900;
            box-shadow: 0 5px 15px rgba(220, 20, 60, 0.5);
            font-family: 'Orbitron', sans-serif;
            animation: pulse 2s infinite;
            z-index: 2;
        }

        .panel-title {
            font-size: 1.8rem;
            margin-bottom: 20px;
            color: var(--primary-red);
            position: relative;
            display: inline-block;
        }

        .panel-title::after {
            content: '';
            position: absolute;
            bottom: 0;
            left: 0;
            width: 100%;
            height: 3px;
            background: linear-gradient(90deg, var(--primary-red), var(--accent-gold));
            border-radius: 2px;
        }

        .panel-content {
            font-size: 1.1rem;
            color: var(--light-text);
            line-height: 1.6;
        }

        .panel-icon {
            font-size: 3rem;
            color: var(--primary-red);
            margin-bottom: 20px;
            display: inline-block;
            animation: float 3s infinite ease-in-out;
        }

        /* Features Grid */
        .features-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 30px;
            margin-top: 50px;
        }

        .feature-card {
            background-color: var(--card-bg);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 25px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            position: relative;
            border: 2px solid var(--primary-red);
            transition: all 0.5s;
            overflow: visible;
            z-index: 1;
            display: block;
        }

        .feature-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 5px;
            background: linear-gradient(90deg, var(--primary-red), var(--accent-gold));
            z-index: -1;
            border-radius: 3px;
        }

        .feature-card:hover {
            transform: translateY(-15px) scale(1.03);
            box-shadow: 0 20px 40px rgba(220, 20, 60, 0.4);
            border-color: var(--accent-gold);
        }

        .feature-icon {
            font-size: 3rem;
            margin-bottom: 15px;
            display: inline-block;
            transition: all 0.3s;
        }

        .feature-card:nth-child(1) .feature-icon { color: var(--primary-red); }
        .feature-card:nth-child(2) .feature-icon { color: var(--accent-gold); }
        .feature-card:nth-child(3) .feature-icon { color: #10B981; }
        .feature-card:nth-child(4) .feature-icon { color: #3B82F6; }
        .feature-card:nth-child(5) .feature-icon { color: #8B5CF6; }
        .feature-card:nth-child(6) .feature-icon { color: #EC4899; }

        .feature-card:hover .feature-icon {
            transform: scale(1.2) rotate(10deg);
        }

        .feature-title {
            font-size: 1.4rem;
            margin-bottom: 15px;
            color: var(--primary-red);
        }

        .feature-description {
            font-size: 1rem;
            color: var(--light-text);
            line-height: 1.6;
        }

        /* How It Works */
        .how-it-works {
            padding: 80px 0;
            background: linear-gradient(135deg, var(--navy-blue) 0%, var(--light-blue) 100%);
            position: relative;
        }

        .steps-container {
            position: relative;
            margin-top: 50px;
        }

        .step-line {
            position: absolute;
            top: 50px;
            left: 50%;
            transform: translateX(-50%);
            width: 4px;
            height: calc(100% - 50px);
            background: linear-gradient(180deg, var(--primary-red), var(--accent-gold));
            z-index: 1;
            box-shadow: 0 0 10px var(--primary-red);
        }

        .step {
            position: relative;
            margin-bottom: 50px;
            z-index: 2;
        }

        .step:nth-child(odd) {
            text-align: right;
            padding-right: calc(50% + 40px);
        }

        .step:nth-child(even) {
            text-align: left;
            padding-left: calc(50% + 40px);
        }

        .step-content {
            background-color: var(--card-bg);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 25px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            border: 3px solid var(--primary-red);
            position: relative;
            transition: all 0.5s;
            display: block;
        }

        .step-content:hover {
            transform: scale(1.05);
            box-shadow: 0 20px 40px rgba(220, 20, 60, 0.4);
            border-color: var(--accent-gold);
        }

        .step-number {
            position: absolute;
            top: -25px;
            width: 60px;
            height: 60px;
            background-color: var(--primary-red);
            color: var(--light-text);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1.8rem;
            font-weight: 900;
            box-shadow: 0 5px 15px rgba(220, 20, 60, 0.5);
            font-family: 'Orbitron', sans-serif;
            animation: pulse 2s infinite;
            z-index: 2;
        }

        .step:nth-child(odd) .step-number { right: -30px; }
        .step:nth-child(even) .step-number { left: -30px; }

        .step-title {
            font-size: 1.5rem;
            margin-bottom: 15px;
            color: var(--primary-red);
        }

        .step-description {
            font-size: 1rem;
            color: var(--light-text);
            line-height: 1.6;
        }

        /* Testimonials */
        .testimonials-section {
            padding: 80px 0;
            background: linear-gradient(180deg, var(--light-blue) 0%, var(--navy-blue) 100%);
            position: relative;
        }

        .testimonial-card {
            background-color: var(--card-bg);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            position: relative;
            border: 3px solid var(--primary-red);
            transition: all 0.5s;
            height: 100%;
            display: block;
        }

        .testimonial-card:hover {
            transform: translateY(-15px) rotate(1deg);
            box-shadow: 0 20px 40px rgba(220, 20, 60, 0.4);
            border-color: var(--accent-gold);
        }

        .quote-icon {
            font-size: 3rem;
            color: var(--primary-red);
            margin-bottom: 20px;
            opacity: 0.7;
        }

        .testimonial-text {
            font-size: 1.1rem;
            color: var(--light-text);
            line-height: 1.6;
            margin-bottom: 20px;
            font-style: italic;
        }

        .testimonial-author {
            font-size: 1.2rem;
            font-weight: 700;
            color: var(--primary-red);
        }

        .testimonial-position {
            font-size: 1rem;
            color: var(--accent-gold);
        }

        /* CTA Section */
        .cta-section {
            padding: 100px 0;
            background: linear-gradient(135deg, var(--primary-red) 0%, #8B0000 100%);
            color: white;
            text-align: center;
            position: relative;
            overflow: hidden;
        }

        .cta-grid {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-image: linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
            background-size: 30px 30px;
            z-index: 0;
            animation: grid-move 15s linear infinite;
        }

        .cta-content {
            position: relative;
            z-index: 1;
        }

        .cta-title {
            font-size: 3rem;
            font-weight: 900;
            margin-bottom: 20px;
            color: var(--light-text);
            text-shadow: 0 0 15px rgba(255, 255, 255, 0.5);
            animation: zoomIn 1s ease-out;
        }

        .cta-subtitle {
            font-size: 1.4rem;
            margin-bottom: 40px;
            color: var(--light-text);
            animation: slideInUp 1s ease-out;
        }

        .btn-light {
            background-color: var(--light-text);
            color: var(--primary-red);
            font-weight: 900;
            padding: 15px 40px;
            border-radius: 50px;
            transition: all 0.3s;
            box-shadow: 0 10px 30px rgba(255, 255, 255, 0.3);
            font-size: 1.2rem;
            border: none;
            text-transform: uppercase;
            letter-spacing: 1px;
            position: relative;
            overflow: hidden;
            z-index: 1;
        }

        .btn-light::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 0;
            height: 100%;
            background-color: var(--accent-gold);
            z-index: -1;
            transition: width 0.3s ease;
        }

        .btn-light:hover::before {
            width: 100%;
        }

        .btn-light:hover {
            transform: translateY(-5px);
            box-shadow: 0 15px 40px rgba(255, 255, 255, 0.5);
        }

        /* Footer */
        footer {
            background-color: var(--navy-blue);
            color: var(--light-text);
            padding: 50px 0 20px;
            border-top: 3px solid var(--primary-red);
        }

        .footer-logo {
            font-weight: 900;
            font-size: 1.8rem;
            margin-bottom: 20px;
            color: var(--primary-red);
            text-shadow: 0 0 10px var(--primary-red);
        }

        .footer-links {
            list-style: none;
            padding: 0;
        }

        .footer-links li {
            margin-bottom: 10px;
        }

        .footer-links a {
            color: var(--light-text);
            text-decoration: none;
            transition: all 0.3s;
            font-weight: 600;
            position: relative;
            display: inline-block;
        }

        .footer-links a::after {
            content: '';
            position: absolute;
            bottom: 0;
            left: 0;
            width: 0;
            height: 2px;
            background-color: var(--primary-red);
            transition: width 0.3s ease;
        }

        .footer-links a:hover::after {
            width: 100%;
        }

        .footer-links a:hover {
            color: var(--primary-red);
            transform: translateX(5px);
        }

        .social-icons a {
            color: var(--light-text);
            font-size: 1.8rem;
            margin: 0 10px;
            transition: all 0.3s;
            display: inline-block;
        }

        .social-icons a:hover {
            color: var(--primary-red);
            transform: translateY(-5px) rotate(15deg);
            text-shadow: 0 0 10px var(--primary-red);
        }

        .copyright {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 2px solid var(--light-blue);
            text-align: center;
            color: var(--light-text);
            opacity: 0.7;
        }

        /* Responsive */
        @media (max-width: 992px) {
            .step:nth-child(odd), .step:nth-child(even) {
                text-align: left;
                padding-left: 70px;
                padding-right: 20px;
            }
            .step-number {
                left: 35px !important;
                right: auto !important;
            }
            .step-line {
                left: 35px;
            }
        }

        @media (max-width: 768px) {
            .hero-title {
                font-size: 2.5rem;
            }
            .section-title {
                font-size: 2rem;
            }
            .cta-title {
                font-size: 2rem;
            }
            .coo-image {
                max-width: 100%;
            }
            .speech-bubble {
                position: static;
                margin-top: 20px;
                max-width: 100%;
            }
            .speech-bubble::after {
                display: none;
            }
        }
    </style>
</head>

<body>
    <!-- Particles Background -->
    <div id="particles-js"></div>

    <!-- Navigation -->
    <nav class="navbar navbar-expand-lg navbar-light fixed-top">
        <div class="container">
            <a class="navbar-brand" href="#">
                <img src="/img/logo.png" alt="COO-LLM Logo" class="brand-logo">
                COO-LLM
            </a>
            <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
                <span class="navbar-toggler-icon"></span>
            </button>
            <div class="collapse navbar-collapse" id="navbarNav">
                <ul class="navbar-nav ms-auto">
                    <li class="nav-item">
                        <a class="nav-link" href="#features">Features</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" href="#how-it-works">How It Works</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" href="#testimonials">Testimonials</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link btn btn-primary ms-2" href="#">Get Started</a>
                    </li>
                </ul>
            </div>
        </div>
    </nav>

    <!-- Hero Section -->
    <section class="hero-section">
        <div class="hero-grid"></div>
        <div class="container">
            <div class="row align-items-center">
                <div class="col-lg-6">
                    <h1 class="hero-title glitch" data-text="I'M YOUR LLM COO">I'M YOUR LLM COO</h1>
                    <p class="hero-subtitle typewriter-container">As your Chief Operations Officer for Large Language Models, I optimize, balance, and streamline all your LLM operations for maximum efficiency and cost savings.</p>
                    <div class="d-flex flex-wrap gap-3 mt-4">
                        <a href="#" class="btn btn-primary">Let's Optimize</a>
                        <a href="#" class="btn btn-outline-primary">See My Strategy</a>
                    </div>
                </div>
                <div class="col-lg-6 mt-5 mt-lg-0">
                    <div class="coo-character">
                        <img src="/img/logo-noback.png" alt="COO-LLM Character" class="coo-image">
                        <div class="speech-bubble">
                            <p class="speech-text">Leave it to me! I'll handle all your LLM operations.</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <!-- COO Introduction -->
    <section class="comic-section">
        <div class="container">
            <div class="text-center mb-5">
                <h2 class="section-title">My Operational Excellence</h2>
            </div>
            <div class="row">
                <div class="col-lg-6 mb-4">
                    <div class="comic-panel">
                        <div class="panel-number">1</div>
                        <h3 class="panel-title">Strategic Resource Allocation</h3>
                        <p class="panel-content">As your COO, I implement intelligent load balancing algorithms that strategically distribute your LLM requests across multiple providers. This ensures optimal resource utilization and prevents bottlenecks in your operations.</p>
                    </div>
                </div>
                <div class="col-lg-6 mb-4">
                    <div class="comic-panel">
                        <div class="panel-number">2</div>
                        <h3 class="panel-title">Seamless Integration</h3>
                        <p class="panel-content">My team and I have engineered a solution with complete OpenAI API compatibility. This means zero code changes for your development team, saving valuable time and resources during implementation.</p>
                    </div>
                </div>
                <div class="col-lg-6 mb-4">
                    <div class="comic-panel">
                        <div class="panel-number">3</div>
                        <h3 class="panel-title">Cost Optimization</h3>
                        <p class="panel-content">I continuously monitor pricing across all providers and route requests to the most cost-effective options in real-time. My algorithms have reduced operational costs by up to 40% for our current clients.</p>
                    </div>
                </div>
                <div class="col-lg-6 mb-4">
                    <div class="comic-panel">
                        <div class="panel-number">4</div>
                        <h3 class="panel-title">Operational Visibility</h3>
                        <p class="panel-content">With enterprise-grade observability tools, I provide comprehensive analytics and reporting on your LLM operations. This data-driven approach enables informed decision-making and continuous improvement.</p>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <!-- Features Section -->
    <section id="features" class="py-5">
        <div class="container">
            <div class="text-center mb-5">
                <h2 class="section-title">My Operational Toolkit</h2>
                <p class="lead tech-font">These are the tools I use to optimize your LLM operations</p>
            </div>
            <div class="features-grid">
                <div class="feature-card">
                    <div class="feature-icon">
                        <i class="fas fa-balance-scale"></i>
                    </div>
                    <h3 class="feature-title">Advanced Load Balancing</h3>
                    <p class="feature-description">My proprietary algorithms distribute requests based on provider capacity, response times, and your specific business priorities.</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <i class="fas fa-plug"></i>
                    </div>
                    <h3 class="feature-title">API Compatibility Layer</h3>
                    <p class="feature-description">Seamlessly switch between providers without modifying your existing codebase. My team handles all the integration complexity.</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <i class="fas fa-chart-line"></i>
                    </div>
                    <h3 class="feature-title">Real-time Cost Analytics</h3>
                    <p class="feature-description">Monitor your spending across all providers with detailed breakdowns and cost-saving recommendations from my analysis team.</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <i class="fas fa-shield-alt"></i>
                    </div>
                    <h3 class="feature-title">Operational Resilience</h3>
                    <p class="feature-description">My failover mechanisms ensure your operations continue smoothly even when individual providers experience downtime or performance issues.</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <i class="fas fa-tachometer-alt"></i>
                    </div>
                    <h3 class="feature-title">Performance Optimization</h3>
                    <p class="feature-description">I continuously benchmark provider performance and route requests to ensure the fastest response times for your critical operations.</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <i class="fas fa-cogs"></i>
                    </div>
                    <h3 class="feature-title">Operational Control</h3>
                    <p class="feature-description">My intuitive dashboard gives you complete control over your LLM operations with customizable routing rules and provider preferences.</p>
                </div>
            </div>
        </div>
    </section>

    <!-- How It Works -->
    <section id="how-it-works" class="how-it-works">
        <div class="container">
            <div class="text-center mb-5">
                <h2 class="section-title">My Implementation Process</h2>
                <p class="lead tech-font">Here's how I streamline your LLM operations in four strategic phases</p>
            </div>
            <div class="steps-container">
                <div class="step-line"></div>
                <div class="step">
                    <div class="step-content">
                        <div class="step-number">1</div>
                        <h3 class="step-title">Assessment & Planning</h3>
                        <p class="step-description">My team begins with a comprehensive assessment of your current LLM usage, costs, and performance requirements. We then develop a customized operational strategy tailored to your business objectives.</p>
                    </div>
                </div>
                <div class="step">
                    <div class="step-content">
                        <div class="step-number">2</div>
                        <h3 class="step-title">Integration & Configuration</h3>
                        <p class="step-description">We handle all technical aspects of integrating your existing systems with COO-LLM. This includes configuring provider connections, API keys, and establishing your operational parameters.</p>
                    </div>
                </div>
                <div class="step">
                    <div class="step-content">
                        <div class="step-number">3</div>
                        <h3 class="step-title">Deployment & Testing</h3>
                        <p class="step-description">My team implements the solution in a phased approach, with rigorous testing at each stage to ensure operational continuity and performance optimization before full deployment.</p>
                    </div>
                </div>
                <div class="step">
                    <div class="step-content">
                        <div class="step-number">4</div>
                        <h3 class="step-title">Optimization & Scaling</h3>
                        <p class="step-description">Post-deployment, I continuously monitor performance metrics and cost data, making iterative improvements to your operational configuration as your business needs evolve.</p>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <!-- Testimonials -->
    <section id="testimonials" class="testimonials-section">
        <div class="container">
            <div class="text-center mb-5">
                <h2 class="section-title">What My Clients Say</h2>
                <p class="lead tech-font">Hear from businesses that have transformed their LLM operations</p>
            </div>
            <div class="row">
                <div class="col-lg-4 mb-4">
                    <div class="testimonial-card">
                        <div class="quote-icon">
                            <i class="fas fa-quote-left"></i>
                        </div>
                        <p class="testimonial-text">"COO-LLM reduced our LLM operational costs by 35% while improving response times. It's like having an expert operations team working 24/7."</p>
                        <h4 class="testimonial-author">Sarah Johnson</h4>
                        <p class="testimonial-position">CTO, TechInnovate</p>
                    </div>
                </div>
                <div class="col-lg-4 mb-4">
                    <div class="testimonial-card">
                        <div class="quote-icon">
                            <i class="fas fa-quote-left"></i>
                        </div>
                        <p class="testimonial-text">"The seamless integration meant zero downtime during implementation. Our development team was amazed at how simple the transition was."</p>
                        <h4 class="testimonial-author">Michael Chen</h4>
                        <p class="testimonial-position">VP Engineering, DataFlow</p>
                    </div>
                </div>
                <div class="col-lg-4 mb-4">
                    <div class="testimonial-card">
                        <div class="quote-icon">
                            <i class="fas fa-quote-left"></i>
                        </div>
                        <p class="testimonial-text">"The visibility and control we now have over our LLM operations is unprecedented. COO-LLM has become an indispensable part of our tech stack."</p>
                        <h4 class="testimonial-author">Emma Rodriguez</h4>
                        <p class="testimonial-position">Director of Operations, CloudScale</p>
                    </div>
                </div>
            </div>
        </div>
    </section>

    <!-- CTA Section -->
    <section class="cta-section">
        <div class="cta-grid"></div>
        <div class="container">
            <div class="cta-content">
                <h2 class="cta-title">Ready to Optimize Your LLM Operations?</h2>
                <p class="cta-subtitle">Let's schedule a consultation to discuss how I can streamline your LLM operations and drive cost savings for your organization.</p>
                <a href="#" class="btn btn-light">Schedule a Strategy Session</a>
            </div>
        </div>
    </section>

    <!-- Footer -->
    <footer>
        <div class="container">
            <div class="row">
                <div class="col-lg-4 mb-4">
                    <div class="footer-logo">COO-LLM</div>
                    <p>Your Chief Operations Officer for Language Models - optimizing, balancing, and streamlining your LLM operations.</p>
                    <div class="social-icons">
                        <a href="#"><i class="fab fa-twitter"></i></a>
                        <a href="#"><i class="fab fa-linkedin"></i></a>
                        <a href="#"><i class="fab fa-github"></i></a>
                        <a href="#"><i class="fas fa-envelope"></i></a>
                    </div>
                </div>
                <div class="col-lg-2 col-md-6 mb-4">
                    <h5>Product</h5>
                    <ul class="footer-links">
                        <li><a href="#">Features</a></li>
                        <li><a href="#">Pricing</a></li>
                        <li><a href="#">Documentation</a></li>
                        <li><a href="#">API Reference</a></li>
                    </ul>
                </div>
                <div class="col-lg-2 col-md-6 mb-4">
                    <h5>Company</h5>
                    <ul class="footer-links">
                        <li><a href="#">About Us</a></li>
                        <li><a href="#">Blog</a></li>
                        <li><a href="#">Careers</a></li>
                        <li><a href="#">Contact</a></li>
                    </ul>
                </div>
                <div class="col-lg-2 col-md-6 mb-4">
                    <h5>Resources</h5>
                    <ul class="footer-links">
                        <li><a href="#">Case Studies</a></li>
                        <li><a href="#">Whitepapers</a></li>
                        <li><a href="#">Webinars</a></li>
                        <li><a href="#">Support</a></li>
                    </ul>
                </div>
                <div class="col-lg-2 col-md-6 mb-4">
                    <h5>Legal</h5>
                    <ul class="footer-links">
                        <li><a href="#">Privacy Policy</a></li>
                        <li><a href="#">Terms of Service</a></li>
                        <li><a href="#">Security</a></li>
                        <li><a href="#">Compliance</a></li>
                    </ul>
                </div>
            </div>
            <div class="copyright">
                <p>&copy; 2023 COO-LLM. All rights reserved. | Optimizing LLM operations worldwide.</p>
            </div>
        </div>
    </footer>

    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>

    <!-- Particles.js -->
    <script src="https://cdn.jsdelivr.net/particles.js/2.0.0/particles.min.js"></script>

    <script>
        // Particles.js Configuration
        particlesJS('particles-js', {
            "particles": {
                "number": {
                    "value": 80,
                    "density": {
                        "enable": true,
                        "value_area": 800
                    }
                },
                "color": {
                    "value": "#DC143C"
                },
                "shape": {
                    "type": "circle",
                    "stroke": {
                        "width": 0,
                        "color": "#000000"
                    }
                },
                "opacity": {
                    "value": 0.5,
                    "random": false,
                    "anim": {
                        "enable": false,
                        "speed": 1,
                        "opacity_min": 0.1,
                        "sync": false
                    }
                },
                "size": {
                    "value": 3,
                    "random": true,
                    "anim": {
                        "enable": false,
                        "speed": 40,
                        "size_min": 0.1,
                        "sync": false
                    }
                },
                "line_linked": {
                    "enable": true,
                    "distance": 150,
                    "color": "#DC143C",
                    "opacity": 0.4,
                    "width": 1
                },
                "move": {
                    "enable": true,
                    "speed": 2,
                    "direction": "none",
                    "random": false,
                    "straight": false,
                    "out_mode": "out",
                    "bounce": false,
                    "attract": {
                        "enable": false,
                        "rotateX": 600,
                        "rotateY": 1200
                    }
                }
            },
            "interactivity": {
                "detect_on": "canvas",
                "events": {
                    "onhover": {
                        "enable": true,
                        "mode": "grab"
                    },
                    "onclick": {
                        "enable": true,
                        "mode": "push"
                    },
                    "resize": true
                },
                "modes": {
                    "grab": {
                        "distance": 140,
                        "line_linked": {
                            "opacity": 1
                        }
                    },
                    "bubble": {
                        "distance": 400,
                        "size": 40,
                        "duration": 2,
                        "opacity": 8,
                        "speed": 3
                    },
                    "repulse": {
                        "distance": 200,
                        "duration": 0.4
                    },
                    "push": {
                        "particles_nb": 4
                    },
                    "remove": {
                        "particles_nb": 2
                    }
                }
            },
            "retina_detect": true
        });

        // Smooth scrolling for navigation links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();

                const targetId = this.getAttribute('href');
                if (targetId === '#') return;

                const targetElement = document.querySelector(targetId);
                if (targetElement) {
                    window.scrollTo({
                        top: targetElement.offsetTop - 80,
                        behavior: 'smooth'
                    });
                }
            });
        });

        // Add shadow to navbar on scroll
        window.addEventListener('scroll', function () {
            const navbar = document.querySelector('.navbar');
            if (window.scrollY > 50) {
                navbar.style.boxShadow = '0 6px 20px rgba(220, 20, 60, 0.5)';
            } else {
                navbar.style.boxShadow = '0 4px 15px rgba(220, 20, 60, 0.3)';
            }
        });

        // Add animation to elements when they come into view
        const observerOptions = {
            root: null,
            rootMargin: '0px',
            threshold: 0.1
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    if (entry.target.classList.contains('comic-panel')) {
                        entry.target.style.animation = 'slideInLeft 0.8s ease-out forwards';
                    } else if (entry.target.classList.contains('feature-card')) {
                        entry.target.style.animation = 'slideInUp 0.8s ease-out forwards';
                    } else if (entry.target.classList.contains('testimonial-card')) {
                        entry.target.style.animation = 'zoomIn 0.8s ease-out forwards';
                    }
                }
            });
        }, observerOptions);

        // Observe all comic panels, feature cards, and testimonial cards
        document.querySelectorAll('.comic-panel, .feature-card, .testimonial-card').forEach(el => {
            observer.observe(el);
        });

        // Add interactive effects to buttons
        document.querySelectorAll('.btn').forEach(button => {
            button.addEventListener('mouseenter', function () {
                this.style.transform = 'translateY(-5px)';
            });

            button.addEventListener('mouseleave', function () {
                this.style.transform = 'translateY(0)';
            });
        });

        // Add ripple effect to buttons
        document.querySelectorAll('.btn').forEach(button => {
            button.addEventListener('click', function (e) {
                const ripple = document.createElement('span');
                const rect = this.getBoundingClientRect();
                const size = Math.max(rect.width, rect.height);
                const x = e.clientX - rect.left - size / 2;
                const y = e.clientY - rect.top - size / 2;

                ripple.style.width = ripple.style.height = size + 'px';
                ripple.style.left = x + 'px';
                ripple.style.top = y + 'px';
                ripple.classList.add('ripple');

                this.appendChild(ripple);

                setTimeout(() => {
                    ripple.remove();
                }, 600);
            });
        });
    </script>

    <style>
        /* Ripple Effect */
        .ripple {
            position: absolute;
            border-radius: 50%;
            background-color: rgba(255, 255, 255, 0.7);
            transform: scale(0);
            animation: ripple-animation 0.6s ease-out;
            pointer-events: none;
        }

        @keyframes ripple-animation {
            to {
                transform: scale(4);
                opacity: 0;
            }
        }

        /* Add position relative to buttons for ripple effect */
        .btn {
            position: relative;
            overflow: hidden;
        }
    </style>
</body>

</html>
` }});
}