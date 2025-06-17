#!/usr/bin/env python3
"""
SMLGOAPI AI Agent Integration Example (Simplified)
===================================================

Simple demonstration of how an AI agent can use the /guide endpoint
to automatically discover and interact with the SMLGOAPI.
"""

import requests
import json

def test_api_guide():
    """Test the /guide endpoint and demonstrate AI agent capabilities"""
    base_url = "http://localhost:8008"
    
    print("🤖 SMLGOAPI AI Agent Simple Test")
    print("=" * 40)
    
    try:
        # Step 1: Get API guide
        print("🔍 Fetching API guide...")
        response = requests.get(f"{base_url}/guide")
        guide = response.json()
        
        print(f"✅ API Name: {guide.get('api_name')}")
        print(f"✅ Version: {guide.get('version')}")
        print(f"✅ Available endpoints: {list(guide.get('endpoints', {}).keys())}")
        print()
        
        # Step 2: Health check
        print("🏥 Checking health...")
        health_response = requests.get(f"{base_url}/health")
        health = health_response.json()
        print(f"✅ Status: {health.get('status')}")
        print()
        
        # Step 3: Test select query
        print("🔍 Testing SELECT query...")
        select_payload = {"query": "SELECT 1 as test, 'AI Agent Test' as message"}
        select_response = requests.post(
            f"{base_url}/select",
            json=select_payload,
            headers={"Content-Type": "application/json"}
        )
        select_result = select_response.json()
        
        if select_result.get('success'):
            print(f"✅ Query successful: {select_result.get('data')}")
        else:
            print(f"❌ Query failed: {select_result.get('error')}")
        print()
        
        # Step 4: Test command
        print("💻 Testing command...")
        command_payload = {"query": "SHOW TABLES"}
        command_response = requests.post(
            f"{base_url}/command",
            json=command_payload,
            headers={"Content-Type": "application/json"}
        )
        command_result = command_response.json()
        
        if command_result.get('success'):
            print(f"✅ Command successful")
        else:
            print(f"❌ Command failed: {command_result.get('error')}")
        print()
        
        print("🎉 AI Agent test completed successfully!")
        
    except Exception as e:
        print(f"❌ Error: {e}")

if __name__ == "__main__":
    test_api_guide()
