#!/usr/bin/env python3
"""
SMLGOAPI Thai Administrative Data Demo
=====================================

This script demonstrates how to use the Thai administrative data API endpoints:
- /get/provinces - Get all provinces
- /get/amphures - Get districts by province
- /get/tambons - Get sub-districts by district and province

Usage Example:
    python thai_admin_demo.py
"""

import requests
import json
from typing import List, Dict, Any

class ThaiAdminAPI:
    def __init__(self, base_url: str = "http://localhost:8008"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json'
        })
    
    def get_provinces(self) -> List[Dict[str, Any]]:
        """Get all Thai provinces"""
        try:
            response = self.session.post(f"{self.base_url}/get/provinces", json={})
            response.raise_for_status()
            data = response.json()
            
            if data.get('success'):
                print(f"✅ {data.get('message')}")
                return data.get('data', [])
            else:
                print(f"❌ Error: {data.get('error')}")
                return []
        except Exception as e:
            print(f"❌ Request failed: {e}")
            return []
    
    def get_amphures(self, province_id: int) -> List[Dict[str, Any]]:
        """Get all amphures in a province"""
        try:
            payload = {"province_id": province_id}
            response = self.session.post(f"{self.base_url}/get/amphures", json=payload)
            response.raise_for_status()
            data = response.json()
            
            if data.get('success'):
                print(f"✅ {data.get('message')}")
                return data.get('data', [])
            else:
                print(f"❌ Error: {data.get('error')}")
                return []
        except Exception as e:
            print(f"❌ Request failed: {e}")
            return []
    
    def get_tambons(self, amphure_id: int, province_id: int) -> List[Dict[str, Any]]:
        """Get all tambons in an amphure"""
        try:
            payload = {"amphure_id": amphure_id, "province_id": province_id}
            response = self.session.post(f"{self.base_url}/get/tambons", json=payload)
            response.raise_for_status()
            data = response.json()
            
            if data.get('success'):
                print(f"✅ {data.get('message')}")
                return data.get('data', [])
            else:
                print(f"❌ Error: {data.get('error')}")
                return []
        except Exception as e:
            print(f"❌ Request failed: {e}")
            return []

def demo_hierarchical_lookup():
    """Demonstrate hierarchical address lookup"""
    print("🏛️  SMLGOAPI Thai Administrative Data Demo")
    print("=" * 50)
    
    api = ThaiAdminAPI()
    
    # 1. Get all provinces
    print("\n🇹🇭 1. Getting all Thai provinces...")
    provinces = api.get_provinces()
    
    if not provinces:
        print("Failed to get provinces. Exiting.")
        return
    
    # Show first 5 provinces
    print("\n📋 First 5 provinces:")
    for province in provinces[:5]:
        print(f"   {province['id']:2d}. {province['name_th']} ({province['name_en']})")
    
    print(f"\n   ... and {len(provinces) - 5} more provinces")
    
    # 2. Get amphures for Bangkok (province_id: 1)
    print("\n🏙️  2. Getting districts in Bangkok...")
    bangkok_amphures = api.get_amphures(1)
    
    if bangkok_amphures:
        print("\n📋 First 5 districts in Bangkok:")
        for amphure in bangkok_amphures[:5]:
            print(f"   {amphure['id']}. {amphure['name_th']} ({amphure['name_en']})")
        
        print(f"\n   ... and {len(bangkok_amphures) - 5} more districts")
        
        # 3. Get tambons for Khet Phra Nakhon (amphure_id: 1001)
        print("\n🏛️  3. Getting sub-districts in Khet Phra Nakhon...")
        tambons = api.get_tambons(1001, 1)
        
        if tambons:
            print("\n📋 All sub-districts in Khet Phra Nakhon:")
            for tambon in tambons:
                print(f"   {tambon['id']}. {tambon['name_th']} ({tambon['name_en']})")
    
    # 4. Example with Chiang Mai
    print("\n🏔️  4. Getting districts in Chiang Mai...")
    chiangmai_amphures = api.get_amphures(38)
    
    if chiangmai_amphures:
        print(f"\n📋 Found {len(chiangmai_amphures)} districts in Chiang Mai:")
        for amphure in chiangmai_amphures[:3]:
            print(f"   {amphure['id']}. {amphure['name_th']} ({amphure['name_en']})")
        
        if len(chiangmai_amphures) > 3:
            print(f"   ... and {len(chiangmai_amphures) - 3} more")

def demo_search_by_name():
    """Demonstrate searching provinces by name"""
    print("\n🔍 Province Search Demo")
    print("-" * 30)
    
    api = ThaiAdminAPI()
    provinces = api.get_provinces()
    
    if not provinces:
        return
    
    # Search examples
    search_terms = ["กรุง", "เชียง", "สมุทร", "นคร"]
    
    for term in search_terms:
        print(f"\n🔍 Provinces containing '{term}':")
        matches = [p for p in provinces if term in p['name_th']]
        
        for match in matches:
            print(f"   {match['id']:2d}. {match['name_th']} ({match['name_en']})")

def main():
    """Main demo function"""
    try:
        # Test connectivity first
        response = requests.get("http://localhost:8008/health", timeout=5)
        if response.status_code != 200:
            print("❌ SMLGOAPI server not available. Please start the server first.")
            return
    except Exception:
        print("❌ Cannot connect to SMLGOAPI server at http://localhost:8008")
        print("   Please make sure the server is running.")
        return
    
    # Run demos
    demo_hierarchical_lookup()
    demo_search_by_name()
    
    print("\n" + "=" * 50)
    print("✅ Demo completed successfully!")
    print("\n💡 Integration Tips:")
    print("   - Use province_id from /get/provinces to get amphures")
    print("   - Use amphure_id + province_id to get tambons")
    print("   - Cache the data locally for better performance")
    print("   - All endpoints return Thai and English names")
    print("   - Perfect for address forms and location selectors")

if __name__ == "__main__":
    main()
